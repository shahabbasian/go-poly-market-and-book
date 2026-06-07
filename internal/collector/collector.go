package collector

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/shahabbasian/polymarket-market-fetcher/internal/api"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/config"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/db"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/models"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/scanner"
)

// Collector continuously fetches deep order books for all active tokens
// using batch CLOB requests and buffered DB writes.
type Collector struct {
	store      *db.PostgresStore
	clobClient *api.CLOBClient
	cfg        *config.Config
	wg         sync.WaitGroup

	// Token cache
	tokenMu     sync.RWMutex
	cachedTokens []models.ActiveToken

	// Snapshot buffer
	bufferMu sync.Mutex
	buffer   []*models.BookSnapshot

	// Last inserted book hash per token_id, used to skip consecutive duplicates
	lastHashMu  sync.Mutex
	lastBookHash map[string]string
}

// New creates a new Collector.
func New(store *db.PostgresStore, clobClient *api.CLOBClient, cfg *config.Config) *Collector {
	return &Collector{
		store:        store,
		clobClient:   clobClient,
		cfg:          cfg,
		buffer:       make([]*models.BookSnapshot, 0, cfg.CollectorBufferMaxRows),
		lastBookHash: make(map[string]string),
	}
}

// Run starts all collector goroutines.
func (c *Collector) Run(ctx context.Context) {
	c.wg.Add(3)
	go c.fetchLoop(ctx)
	go c.flushLoop(ctx)
	go c.tokenRefreshLoop(ctx)
}

// Wait blocks until all collector goroutines have stopped.
func (c *Collector) Wait() {
	c.wg.Wait()
}

// ---- Token cache ----

func (c *Collector) getCachedTokens() []models.ActiveToken {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.cachedTokens
}

func (c *Collector) setCachedTokens(tokens []models.ActiveToken) {
	c.tokenMu.Lock()
	c.cachedTokens = tokens
	c.tokenMu.Unlock()
}

func (c *Collector) tokenRefreshLoop(ctx context.Context) {
	defer c.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("collector token refresh panic", "recover", r)
		}
	}()

	// Initial load
	c.refreshTokens(ctx)

	ticker := time.NewTicker(time.Duration(c.cfg.CollectorTokenRefreshS) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.refreshTokens(ctx)
		}
	}
}

func (c *Collector) refreshTokens(ctx context.Context) {
	// Generate current slugs for all 28 coin×interval combos
	var slugs []string
	for _, coin := range models.Coins {
		for _, iv := range models.Intervals {
			if sl := scanner.CurrentSlug(coin.Symbol, iv.Name); sl != "" {
				slugs = append(slugs, sl)
			}
		}
	}

	tokens, err := c.store.GetActiveMarketTokens(ctx, slugs)
	if err != nil {
		slog.Error("collector failed to refresh tokens", "error", err)
		return
	}
	if len(tokens) == 0 {
		slog.Warn("collector: no active tokens found for current time window")
		return
	}
	c.setCachedTokens(tokens)
	slog.Info("collector refreshed token list", "count", len(tokens))
}

// ---- Fetch loop: batch CLOB requests at 3 req/s ----

func (c *Collector) fetchLoop(ctx context.Context) {
	defer c.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("collector fetch loop panic", "recover", r)
		}
	}()

	delay := time.Duration(c.cfg.CollectorDelayMS) * time.Millisecond
	timer := time.NewTimer(delay)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			c.fetchAndBuffer(ctx)
			timer.Reset(delay)
		}
	}
}

func (c *Collector) fetchAndBuffer(ctx context.Context) {
	tokens := c.getCachedTokens()
	if len(tokens) == 0 {
		return
	}

	start := time.Now()
	totalFetched := 0

	// CLOB API has a payload limit — split into batches of 50 tokens.
	batchSize := c.cfg.CLOBBatchSize
	if batchSize <= 0 || batchSize > 50 {
		batchSize = 50
	}

	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]

		// Build token IDs for this batch
		ids := make([]string, len(batch))
		for j, tok := range batch {
			ids[j] = tok.TokenID
		}

		// Rate limit is handled by the CLOBClient internally.
		books, err := c.clobClient.FetchBooksBatch(ctx, ids)
		if err != nil {
			slog.Warn("collector batch fetch failed", "tokens", len(ids), "error", err)
			return
		}

		// Parse into BookSnapshot
		var snaps []*models.BookSnapshot
		for _, tok := range batch {
			book, ok := books[tok.TokenID]
			if !ok || book == nil {
				continue
			}
			snaps = append(snaps, c.buildSnapshot(tok, book))
		}

		if len(snaps) == 0 {
			continue
		}

		// Append to buffer — enforce hard limit
		c.bufferMu.Lock()
		if len(c.buffer)+len(snaps) > c.cfg.CollectorBufferMaxRows {
			slog.Warn("collector buffer full, dropping snapshots", "buffer", len(c.buffer), "dropping", len(snaps), "max", c.cfg.CollectorBufferMaxRows)
			c.bufferMu.Unlock()
			return
		}
		c.buffer = append(c.buffer, snaps...)
		totalFetched += len(snaps)
		c.bufferMu.Unlock()
	}

	slog.Debug("collector fetched batch",
		"total_fetched", totalFetched,
		"buffer", len(c.buffer),
		"duration", time.Since(start),
	)
}

// ---- Flush loop: batch insert to DB every N seconds ----

func (c *Collector) flushLoop(ctx context.Context) {
	defer c.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("collector flush loop panic", "recover", r)
		}
	}()

	interval := time.Duration(c.cfg.CollectorFlushIntervalS) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Do one final flush on shutdown (with 10s timeout so we don't block forever)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		c.flushBuffer(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.flushBuffer(ctx)
		}
	}
}

func (c *Collector) flushBuffer(ctx context.Context) {
	// Swap buffer under lock to minimise contention
	c.bufferMu.Lock()
	if len(c.buffer) == 0 {
		c.bufferMu.Unlock()
		return
	}
	toFlush := c.buffer
	c.buffer = make([]*models.BookSnapshot, 0, c.cfg.CollectorBufferMaxRows)
	c.bufferMu.Unlock()

	start := time.Now()
	totalRows := len(toFlush)

	// Filter out consecutive duplicate book hashes per token.
	// We keep the first snapshot of a new hash and skip later ones in this
	// batch that match the last known hash.  This prevents DB bloat for
	// stale markets (1h, 4h) without losing legitimate returns to a
	// previous hash after an intervening change.
	var filtered []*models.BookSnapshot
	c.lastHashMu.Lock()
	for _, snap := range toFlush {
		if snap.BookHash != nil {
			if last, ok := c.lastBookHash[snap.TokenID]; ok && last == *snap.BookHash {
				continue // exact duplicate of last inserted state
			}
			c.lastBookHash[snap.TokenID] = *snap.BookHash
		}
		filtered = append(filtered, snap)
	}
	c.lastHashMu.Unlock()

	if len(filtered) == 0 {
		slog.Debug("collector flush skipped, all rows were consecutive duplicates",
			"buffered", totalRows,
		)
		return
	}

	// 1. Insert filtered snapshots into time-series table
	if err := c.store.InsertBookSnapshotsBatch(ctx, filtered); err != nil {
		slog.Error("collector batch insert failed, buffer dropped", "rows", len(filtered), "error", err)
		return
	}

	// 2. For current_books, only keep the latest snapshot per token_id
	latest := c.dedupLatest(filtered)

	if err := c.store.UpsertCurrentBooksBatch(ctx, latest); err != nil {
		slog.Error("collector batch upsert failed", "rows", len(latest), "error", err)
		return
	}

	slog.Info("collector flushed to DB",
		"snapshots", len(filtered),
		"skipped_dups", totalRows-len(filtered),
		"current_books_upserted", len(latest),
		"duration", time.Since(start),
	)
}

// dedupLatest keeps only the most recent snapshot per token_id.
// Iterates in forward order — later duplicates overwrite earlier ones
// in the map, so the last (latest) per token_id wins.
func (c *Collector) dedupLatest(snaps []*models.BookSnapshot) []*models.BookSnapshot {
	latest := make(map[string]*models.BookSnapshot, 56)
	order := make([]string, 0, 56) // track first-appearance order

	for _, s := range snaps {
		if _, exists := latest[s.TokenID]; !exists {
			order = append(order, s.TokenID)
		}
		latest[s.TokenID] = s
	}

	result := make([]*models.BookSnapshot, 0, len(latest))
	for _, id := range order {
		if s, ok := latest[id]; ok {
			result = append(result, s)
		}
	}
	return result
}

// ---- Build snapshot from CLOB response ----

func (c *Collector) buildSnapshot(tok models.ActiveToken, book *models.CLOBBookResponse) *models.BookSnapshot {
	snap := &models.BookSnapshot{
		TokenID:  tok.TokenID,
		Side:     tok.Side,
		Symbol:   tok.Symbol,
		Interval: tok.Interval,
		TimestampAPI: parseRawTimestamp(book.Timestamp),
	}

	// Best bid = bids[0] (sorted descending by price)
	if len(book.Bids) > 0 {
		bid, _ := strconv.ParseFloat(book.Bids[0].Price, 64)
		snap.BestBid = &bid
		sz, _ := strconv.ParseFloat(book.Bids[0].Size, 64)
		snap.BidSize = &sz
	}

	// Best ask = asks[0] (sorted ascending by price)
	if len(book.Asks) > 0 {
		ask, _ := strconv.ParseFloat(book.Asks[0].Price, 64)
		snap.BestAsk = &ask
		sz, _ := strconv.ParseFloat(book.Asks[0].Size, 64)
		snap.AskSize = &sz
	}

	// Spread
	if snap.BestBid != nil && snap.BestAsk != nil {
		spread := *snap.BestAsk - *snap.BestBid
		snap.Spread = &spread
	}

	// Last trade
	if book.LastTradePrice != "" {
		ltp, _ := strconv.ParseFloat(book.LastTradePrice, 64)
		snap.LastTrade = &ltp
	}

	// Book hash
	if book.Hash != "" {
		h := book.Hash
		snap.BookHash = &h
	}

	// Full order book as JSONB (all levels)
	if bidsJSON, err := json.Marshal(book.Bids); err == nil {
		snap.RawBids = bidsJSON
	}
	if asksJSON, err := json.Marshal(book.Asks); err == nil {
		snap.RawAsks = asksJSON
	}

	return snap
}

// parseRawTimestamp returns the raw epoch-millisecond value from the API string.
func parseRawTimestamp(s string) *int64 {
	if s == "" {
		return nil
	}
	if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
		return &ms
	}
	return nil
}
