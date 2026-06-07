package scanner

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shahabbasian/polymarket-market-fetcher/internal/api"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/config"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/db"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/models"
)

// Scanner orchestrates the discovery and refresh workers.
type Scanner struct {
	store       *db.PostgresStore
	gammaClient *api.GammaClient
	clobClient  *api.CLOBClient
	cfg         *config.Config
	wg          sync.WaitGroup
}

func New(store *db.PostgresStore, gammaClient *api.GammaClient, clobClient *api.CLOBClient, cfg *config.Config) *Scanner {
	return &Scanner{
		store:       store,
		gammaClient: gammaClient,
		clobClient:  clobClient,
		cfg:         cfg,
	}
}

func (s *Scanner) Run(ctx context.Context) {
	s.wg.Add(2)
	go s.discoveryWorker(ctx)
	go s.refreshWorker(ctx)
}

func (s *Scanner) Wait() {
	s.wg.Wait()
}

// discoveryWorker continuously finds new markets.
func (s *Scanner) discoveryWorker(ctx context.Context) {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("discovery worker panic", "recover", r)
		}
	}()

	ticker := time.NewTicker(s.cfg.DiscoveryInterval)
	defer ticker.Stop()

	// Run immediately on start
	s.runDiscovery(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("discovery worker shutting down")
			return
		case <-ticker.C:
			s.runDiscovery(ctx)
		}
	}
}

func (s *Scanner) runDiscovery(ctx context.Context) {
	start := time.Now()
	slog.Info("discovery cycle started")

	latestSlugs, err := s.store.GetLatestSlugs(ctx)
	if err != nil {
		slog.Error("failed to get latest slugs", "error", err)
		return
	}

	var tasks []Task
	for _, coin := range models.Coins {
		for _, ivl := range models.Intervals {
			key := coin.Symbol + ":" + ivl.Name
			latestSlug, hasHistory := latestSlugs[key]
			var slugs []string
			var err error
			if hasHistory {
				slugs, err = GenerateNextSlugs(coin.Symbol, ivl.Name, latestSlug, s.cfg.FutureSlugCount)
			} else {
				slugs, err = GenerateInitialSlugs(coin.Symbol, ivl.Name, s.cfg.InitialSlugCount)
			}
			if err != nil {
				slog.Warn("slug generation failed", "symbol", coin.Symbol, "interval", ivl.Name, "error", err)
				continue
			}
			for _, sl := range slugs {
				tasks = append(tasks, Task{
					Symbol:   coin.Symbol,
					Interval: ivl.Name,
					Slug:     sl,
				})
			}
		}
	}

	slog.Info("generated discovery tasks", "count", len(tasks))

	// Fetch Gamma data for each slug
	var markets []*models.NewMarket
	var tokenIDsToFetch []string

	for _, task := range tasks {
		if ctx.Err() != nil {
			return
		}

		market, err := s.gammaClient.FetchMarketBySlug(ctx, task.Slug)
		if err != nil {
			if _, ok := err.(api.NotFoundError); ok {
				slog.Debug("market not found", "slug", task.Slug)
				continue
			}
			slog.Warn("gamma fetch failed", "slug", task.Slug, "error", err)
			continue
		}

		m := s.mapGammaToMarket(market, task.Symbol, task.Interval)
		markets = append(markets, m)

		// Collect token IDs for CLOB batch
		if m.TokenIDYes != nil && *m.TokenIDYes != "" {
			tokenIDsToFetch = append(tokenIDsToFetch, *m.TokenIDYes)
		}
		if m.TokenIDNo != nil && *m.TokenIDNo != "" {
			tokenIDsToFetch = append(tokenIDsToFetch, *m.TokenIDNo)
		}
	}

	// Batch upsert markets (without CLOB data first)
	if len(markets) > 0 {
		if err := s.store.UpsertMarketsBatch(ctx, markets); err != nil {
			slog.Error("batch upsert failed", "error", err)
		}
	}

	// Fetch CLOB books in batches
	if len(tokenIDsToFetch) > 0 {
		s.fetchAndMergeCLOBData(ctx, markets, tokenIDsToFetch)
	}

	slog.Info("discovery cycle completed",
		"duration", time.Since(start),
		"markets_found", len(markets),
		"clob_tokens", len(tokenIDsToFetch),
	)
}

// refreshWorker keeps existing markets up to date.
func (s *Scanner) refreshWorker(ctx context.Context) {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("refresh worker panic", "recover", r)
		}
	}()

	// Initial delay so discovery runs first
	select {
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Second):
	}

	ticker := time.NewTicker(s.cfg.RefreshInterval)
	defer ticker.Stop()

	s.runRefresh(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("refresh worker shutting down")
			return
		case <-ticker.C:
			s.runRefresh(ctx)
		}
	}
}

func (s *Scanner) runRefresh(ctx context.Context) {
	start := time.Now()
	slog.Info("refresh cycle started")

	markets, err := s.store.GetMarketsNeedingRefresh(ctx, s.cfg.RefreshMaxAge, s.cfg.RefreshLimit)
	if err != nil {
		slog.Error("failed to get markets needing refresh", "error", err)
		return
	}

	if len(markets) == 0 {
		slog.Debug("no markets need refresh")
		return
	}

	var updatedMarkets []*models.NewMarket
	var tokenIDsToFetch []string

	for _, m := range markets {
		if ctx.Err() != nil {
			return
		}

		// Re-fetch from Gamma
		gammaMarket, err := s.gammaClient.FetchMarketBySlug(ctx, m.Slug)
		if err != nil {
			if _, ok := err.(api.NotFoundError); ok {
				slog.Debug("market not found on refresh", "slug", m.Slug)
				continue
			}
			slog.Warn("gamma refresh failed", "slug", m.Slug, "error", err)
			continue
		}

		updated := s.mapGammaToMarket(gammaMarket, m.Symbol, m.Interval)
		// Preserve token IDs if Gamma returned them empty
		if updated.TokenIDYes == nil || *updated.TokenIDYes == "" {
			updated.TokenIDYes = m.TokenIDYes
		}
		if updated.TokenIDNo == nil || *updated.TokenIDNo == "" {
			updated.TokenIDNo = m.TokenIDNo
		}
		updated.ID = m.ID
		updated.CreatedAt = m.CreatedAt

		updatedMarkets = append(updatedMarkets, updated)

		if updated.TokenIDYes != nil && *updated.TokenIDYes != "" {
			tokenIDsToFetch = append(tokenIDsToFetch, *updated.TokenIDYes)
		}
		if updated.TokenIDNo != nil && *updated.TokenIDNo != "" {
			tokenIDsToFetch = append(tokenIDsToFetch, *updated.TokenIDNo)
		}
	}

	if len(updatedMarkets) > 0 {
		if err := s.store.UpsertMarketsBatch(ctx, updatedMarkets); err != nil {
			slog.Error("refresh batch upsert failed", "error", err)
		}
	}

	if len(tokenIDsToFetch) > 0 {
		s.fetchAndMergeCLOBData(ctx, updatedMarkets, tokenIDsToFetch)
	}

	slog.Info("refresh cycle completed",
		"duration", time.Since(start),
		"markets_refreshed", len(updatedMarkets),
	)
}

// fetchAndMergeCLOBData batches CLOB book fetches and merges into markets.
func (s *Scanner) fetchAndMergeCLOBData(ctx context.Context, markets []*models.NewMarket, tokenIDs []string) {
	batchSize := s.cfg.CLOBBatchSize
	for i := 0; i < len(tokenIDs); i += batchSize {
		end := i + batchSize
		if end > len(tokenIDs) {
			end = len(tokenIDs)
		}
		batch := tokenIDs[i:end]

		if ctx.Err() != nil {
			return
		}

		books, err := s.clobClient.FetchBooksBatch(ctx, batch)
		if err != nil {
			slog.Warn("clob batch fetch failed", "error", err)
			continue
		}

		// Merge CLOB data into markets
		for _, m := range markets {
			if m.TokenIDYes != nil {
				if book, ok := books[*m.TokenIDYes]; ok {
					s.mergeBookData(m, book)
				}
			}
			if m.TokenIDNo != nil {
				if book, ok := books[*m.TokenIDNo]; ok {
					// Only store hash from no token if yes didn't have one
					if m.LastBookHash == nil || *m.LastBookHash == "" {
						m.LastBookHash = &book.Hash
					}
				}
			}
		}
	}

	// Upsert the updated markets with CLOB data
	if err := s.store.UpsertMarketsBatch(ctx, markets); err != nil {
		slog.Error("clob merge upsert failed", "error", err)
	}
}

func (s *Scanner) mergeBookData(m *models.NewMarket, book *models.CLOBBookResponse) {
	if len(book.Bids) > 0 {
		bid, _ := strconv.ParseFloat(book.Bids[0].Price, 64)
		m.BestBid = &bid
	}
	if len(book.Asks) > 0 {
		ask, _ := strconv.ParseFloat(book.Asks[0].Price, 64)
		m.BestAsk = &ask
	}
	if book.LastTradePrice != "" {
		ltp, _ := strconv.ParseFloat(book.LastTradePrice, 64)
		m.LastTradePrice = &ltp
	}
	if book.Hash != "" {
		m.LastBookHash = &book.Hash
	}
}

// mapGammaToMarket converts a Gamma API response to our DB model.
func (s *Scanner) mapGammaToMarket(g *models.GammaMarketResponse, symbol, interval string) *models.NewMarket {
	m := &models.NewMarket{
		Symbol:   symbol,
		Interval: interval,
		Slug:     g.Slug,
		Status:   "upcoming",
	}

	if g.ConditionID != "" {
		m.ConditionID = &g.ConditionID
	}
	if g.Question != "" {
		m.Question = &g.Question
	}
	if g.Outcomes != "" {
		m.Outcomes = &g.Outcomes
	}
	if g.ID != "" {
		m.GammaMarketID = &g.ID
	}

	// Parse start/end dates
	if g.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, g.StartDate); err == nil {
			m.StartDate = &t
		}
	}
	if g.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, g.EndDate); err == nil {
			m.EndDate = &t
		}
	}

	// Booleans
	m.EnableOrderBook = &g.EnableOrderBook
	m.AcceptingOrders = &g.AcceptingOrders
	m.Ready = &g.Ready
	m.Funded = &g.Funded

	// Numeric fields
	if g.OrderMinSize > 0 {
		m.OrderMinSize = &g.OrderMinSize
	}
	if g.OrderPriceMinTickSize > 0 {
		m.OrderPriceMinTickSize = &g.OrderPriceMinTickSize
	}
	if g.BestBid > 0 {
		m.BestBid = &g.BestBid
	}
	if g.BestAsk > 0 {
		m.BestAsk = &g.BestAsk
	}
	if g.LastTradePrice > 0 {
		m.LastTradePrice = &g.LastTradePrice
	}
	if g.VolumeCLOB > 0 {
		m.VolumeCLOB = &g.VolumeCLOB
	}
	if g.VolumeNum > 0 {
		m.VolumeNum = &g.VolumeNum
	}

	// Status
	if g.Closed {
		m.Status = "closed"
	} else if g.Active && g.Ready && g.Funded {
		m.Status = "active"
	} else if g.Funded {
		m.Status = "funded"
	} else {
		m.Status = "upcoming"
	}

	// Parse clobTokenIds
	if g.ClobTokenIDs != nil && *g.ClobTokenIDs != "" {
		var tokens []string
		if err := json.Unmarshal([]byte(*g.ClobTokenIDs), &tokens); err == nil && len(tokens) >= 2 {
			yesToken := tokens[0]
			noToken := tokens[1]
			m.TokenIDYes = &yesToken
			m.TokenIDNo = &noToken
		} else {
			// Try comma-separated fallback
			parts := strings.Split(*g.ClobTokenIDs, ",")
			if len(parts) >= 2 {
				p0 := strings.TrimSpace(parts[0])
				p1 := strings.TrimSpace(parts[1])
				m.TokenIDYes = &p0
				m.TokenIDNo = &p1
			}
		}
	}

	return m
}
