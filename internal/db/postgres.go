package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/models"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing database URL: %w", err)
	}
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) Pool() *pgxpool.Pool {
	return s.pool
}

// GetLatestSlugs returns the latest slug per (symbol, interval) combination.
func (s *PostgresStore) GetLatestSlugs(ctx context.Context) (map[string]string, error) {
	query := `
		SELECT DISTINCT ON (symbol, interval) symbol, interval, slug
		FROM public.new_markets
		WHERE start_date IS NOT NULL
		ORDER BY symbol, interval, start_date DESC
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying latest slugs: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var sym, interval, slug string
		if err := rows.Scan(&sym, &interval, &slug); err != nil {
			continue
		}
		key := sym + ":" + interval
		result[key] = slug
	}
	return result, rows.Err()
}

// UpsertMarket inserts or updates a market row.
// Uses token_id_yes unique constraint when available; falls back to (symbol, interval, start_date).
func (s *PostgresStore) UpsertMarket(ctx context.Context, m *models.NewMarket) error {
	// Prefer upsert on token_id_yes when it's set
	if m.TokenIDYes != nil && *m.TokenIDYes != "" {
		return s.upsertByTokenIDYes(ctx, m)
	}
	// Fallback: upsert by (symbol, interval, start_date) or slug
	return s.upsertBySlug(ctx, m)
}

func (s *PostgresStore) upsertByTokenIDYes(ctx context.Context, m *models.NewMarket) error {
	query := `
		INSERT INTO public.new_markets (
			symbol, interval, condition_id, token_id_yes, token_id_no,
			question, slug, outcomes, start_date, end_date,
			gamma_market_id, enable_order_book, accepting_orders, ready, funded,
			order_min_size, order_price_min_tick_size,
			best_bid, best_ask, last_trade_price, volume_clob, volume_num,
			status, winning_outcome, price_to_beat, last_book_hash,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22,
			$23, $24, $25, $26,
			now()
		) ON CONFLICT (token_id_yes) WHERE token_id_yes IS NOT NULL DO UPDATE SET
			condition_id = COALESCE(EXCLUDED.condition_id, new_markets.condition_id),
			token_id_no = COALESCE(EXCLUDED.token_id_no, new_markets.token_id_no),
			question = COALESCE(EXCLUDED.question, new_markets.question),
			slug = COALESCE(EXCLUDED.slug, new_markets.slug),
			outcomes = COALESCE(EXCLUDED.outcomes, new_markets.outcomes),
			start_date = COALESCE(EXCLUDED.start_date, new_markets.start_date),
			end_date = COALESCE(EXCLUDED.end_date, new_markets.end_date),
			gamma_market_id = COALESCE(EXCLUDED.gamma_market_id, new_markets.gamma_market_id),
			enable_order_book = COALESCE(EXCLUDED.enable_order_book, new_markets.enable_order_book),
			accepting_orders = COALESCE(EXCLUDED.accepting_orders, new_markets.accepting_orders),
			ready = COALESCE(EXCLUDED.ready, new_markets.ready),
			funded = COALESCE(EXCLUDED.funded, new_markets.funded),
			order_min_size = COALESCE(EXCLUDED.order_min_size, new_markets.order_min_size),
			order_price_min_tick_size = COALESCE(EXCLUDED.order_price_min_tick_size, new_markets.order_price_min_tick_size),
			best_bid = COALESCE(EXCLUDED.best_bid, new_markets.best_bid),
			best_ask = COALESCE(EXCLUDED.best_ask, new_markets.best_ask),
			last_trade_price = COALESCE(EXCLUDED.last_trade_price, new_markets.last_trade_price),
			volume_clob = COALESCE(EXCLUDED.volume_clob, new_markets.volume_clob),
			volume_num = COALESCE(EXCLUDED.volume_num, new_markets.volume_num),
			status = COALESCE(EXCLUDED.status, new_markets.status),
			winning_outcome = COALESCE(EXCLUDED.winning_outcome, new_markets.winning_outcome),
			price_to_beat = COALESCE(EXCLUDED.price_to_beat, new_markets.price_to_beat),
			last_book_hash = COALESCE(EXCLUDED.last_book_hash, new_markets.last_book_hash),
			updated_at = EXCLUDED.updated_at
	`
	_, err := s.pool.Exec(ctx, query,
		m.Symbol, m.Interval, m.ConditionID, m.TokenIDYes, m.TokenIDNo,
		m.Question, m.Slug, m.Outcomes, m.StartDate, m.EndDate,
		m.GammaMarketID, m.EnableOrderBook, m.AcceptingOrders, m.Ready, m.Funded,
		m.OrderMinSize, m.OrderPriceMinTickSize,
		m.BestBid, m.BestAsk, m.LastTradePrice, m.VolumeCLOB, m.VolumeNum,
		m.Status, m.WinningOutcome, m.PriceToBeat, m.LastBookHash,
	)
	if err != nil {
		return fmt.Errorf("upsert by token_id_yes: %w", err)
	}
	return nil
}

func (s *PostgresStore) upsertBySlug(ctx context.Context, m *models.NewMarket) error {
	query := `
		INSERT INTO public.new_markets (
			symbol, interval, condition_id, token_id_yes, token_id_no,
			question, slug, outcomes, start_date, end_date,
			gamma_market_id, enable_order_book, accepting_orders, ready, funded,
			order_min_size, order_price_min_tick_size,
			best_bid, best_ask, last_trade_price, volume_clob, volume_num,
			status, winning_outcome, price_to_beat, last_book_hash,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22,
			$23, $24, $25, $26,
			now()
		) ON CONFLICT (slug) DO UPDATE SET
			condition_id = COALESCE(EXCLUDED.condition_id, new_markets.condition_id),
			token_id_yes = COALESCE(EXCLUDED.token_id_yes, new_markets.token_id_yes),
			token_id_no = COALESCE(EXCLUDED.token_id_no, new_markets.token_id_no),
			question = COALESCE(EXCLUDED.question, new_markets.question),
			outcomes = COALESCE(EXCLUDED.outcomes, new_markets.outcomes),
			start_date = COALESCE(EXCLUDED.start_date, new_markets.start_date),
			end_date = COALESCE(EXCLUDED.end_date, new_markets.end_date),
			gamma_market_id = COALESCE(EXCLUDED.gamma_market_id, new_markets.gamma_market_id),
			enable_order_book = COALESCE(EXCLUDED.enable_order_book, new_markets.enable_order_book),
			accepting_orders = COALESCE(EXCLUDED.accepting_orders, new_markets.accepting_orders),
			ready = COALESCE(EXCLUDED.ready, new_markets.ready),
			funded = COALESCE(EXCLUDED.funded, new_markets.funded),
			order_min_size = COALESCE(EXCLUDED.order_min_size, new_markets.order_min_size),
			order_price_min_tick_size = COALESCE(EXCLUDED.order_price_min_tick_size, new_markets.order_price_min_tick_size),
			best_bid = COALESCE(EXCLUDED.best_bid, new_markets.best_bid),
			best_ask = COALESCE(EXCLUDED.best_ask, new_markets.best_ask),
			last_trade_price = COALESCE(EXCLUDED.last_trade_price, new_markets.last_trade_price),
			volume_clob = COALESCE(EXCLUDED.volume_clob, new_markets.volume_clob),
			volume_num = COALESCE(EXCLUDED.volume_num, new_markets.volume_num),
			status = COALESCE(EXCLUDED.status, new_markets.status),
			winning_outcome = COALESCE(EXCLUDED.winning_outcome, new_markets.winning_outcome),
			price_to_beat = COALESCE(EXCLUDED.price_to_beat, new_markets.price_to_beat),
			last_book_hash = COALESCE(EXCLUDED.last_book_hash, new_markets.last_book_hash),
			updated_at = EXCLUDED.updated_at
	`
	_, err := s.pool.Exec(ctx, query,
		m.Symbol, m.Interval, m.ConditionID, m.TokenIDYes, m.TokenIDNo,
		m.Question, m.Slug, m.Outcomes, m.StartDate, m.EndDate,
		m.GammaMarketID, m.EnableOrderBook, m.AcceptingOrders, m.Ready, m.Funded,
		m.OrderMinSize, m.OrderPriceMinTickSize,
		m.BestBid, m.BestAsk, m.LastTradePrice, m.VolumeCLOB, m.VolumeNum,
		m.Status, m.WinningOutcome, m.PriceToBeat, m.LastBookHash,
	)
	if err != nil {
		return fmt.Errorf("upsert by slug: %w", err)
	}
	return nil
}

// UpsertMarketsBatch upserts multiple markets in a single transaction.
func (s *PostgresStore) UpsertMarketsBatch(ctx context.Context, markets []*models.NewMarket) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, m := range markets {
		if err := s.upsertMarketTx(ctx, tx, m); err != nil {
			slog.Warn("batch upsert failed for one market", "slug", m.Slug, "error", err)
			// Continue with other markets in batch
			continue
		}
	}

	return tx.Commit(ctx)
}

func (s *PostgresStore) upsertMarketTx(ctx context.Context, tx pgx.Tx, m *models.NewMarket) error {
	if m.TokenIDYes != nil && *m.TokenIDYes != "" {
		return s.upsertByTokenIDYesTx(ctx, tx, m)
	}
	return s.upsertBySlugTx(ctx, tx, m)
}

func (s *PostgresStore) upsertByTokenIDYesTx(ctx context.Context, tx pgx.Tx, m *models.NewMarket) error {
	query := `
		INSERT INTO public.new_markets (
			symbol, interval, condition_id, token_id_yes, token_id_no,
			question, slug, outcomes, start_date, end_date,
			gamma_market_id, enable_order_book, accepting_orders, ready, funded,
			order_min_size, order_price_min_tick_size,
			best_bid, best_ask, last_trade_price, volume_clob, volume_num,
			status, winning_outcome, price_to_beat, last_book_hash,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22,
			$23, $24, $25, $26,
			now()
		) ON CONFLICT (token_id_yes) WHERE token_id_yes IS NOT NULL DO UPDATE SET
			condition_id = COALESCE(EXCLUDED.condition_id, new_markets.condition_id),
			token_id_no = COALESCE(EXCLUDED.token_id_no, new_markets.token_id_no),
			question = COALESCE(EXCLUDED.question, new_markets.question),
			slug = COALESCE(EXCLUDED.slug, new_markets.slug),
			outcomes = COALESCE(EXCLUDED.outcomes, new_markets.outcomes),
			start_date = COALESCE(EXCLUDED.start_date, new_markets.start_date),
			end_date = COALESCE(EXCLUDED.end_date, new_markets.end_date),
			gamma_market_id = COALESCE(EXCLUDED.gamma_market_id, new_markets.gamma_market_id),
			enable_order_book = COALESCE(EXCLUDED.enable_order_book, new_markets.enable_order_book),
			accepting_orders = COALESCE(EXCLUDED.accepting_orders, new_markets.accepting_orders),
			ready = COALESCE(EXCLUDED.ready, new_markets.ready),
			funded = COALESCE(EXCLUDED.funded, new_markets.funded),
			order_min_size = COALESCE(EXCLUDED.order_min_size, new_markets.order_min_size),
			order_price_min_tick_size = COALESCE(EXCLUDED.order_price_min_tick_size, new_markets.order_price_min_tick_size),
			best_bid = COALESCE(EXCLUDED.best_bid, new_markets.best_bid),
			best_ask = COALESCE(EXCLUDED.best_ask, new_markets.best_ask),
			last_trade_price = COALESCE(EXCLUDED.last_trade_price, new_markets.last_trade_price),
			volume_clob = COALESCE(EXCLUDED.volume_clob, new_markets.volume_clob),
			volume_num = COALESCE(EXCLUDED.volume_num, new_markets.volume_num),
			status = COALESCE(EXCLUDED.status, new_markets.status),
			winning_outcome = COALESCE(EXCLUDED.winning_outcome, new_markets.winning_outcome),
			price_to_beat = COALESCE(EXCLUDED.price_to_beat, new_markets.price_to_beat),
			last_book_hash = COALESCE(EXCLUDED.last_book_hash, new_markets.last_book_hash),
			updated_at = EXCLUDED.updated_at
	`
	_, err := tx.Exec(ctx, query,
		m.Symbol, m.Interval, m.ConditionID, m.TokenIDYes, m.TokenIDNo,
		m.Question, m.Slug, m.Outcomes, m.StartDate, m.EndDate,
		m.GammaMarketID, m.EnableOrderBook, m.AcceptingOrders, m.Ready, m.Funded,
		m.OrderMinSize, m.OrderPriceMinTickSize,
		m.BestBid, m.BestAsk, m.LastTradePrice, m.VolumeCLOB, m.VolumeNum,
		m.Status, m.WinningOutcome, m.PriceToBeat, m.LastBookHash,
	)
	if err != nil {
		return fmt.Errorf("upsert by token_id_yes tx: %w", err)
	}
	return nil
}

func (s *PostgresStore) upsertBySlugTx(ctx context.Context, tx pgx.Tx, m *models.NewMarket) error {
	query := `
		INSERT INTO public.new_markets (
			symbol, interval, condition_id, token_id_yes, token_id_no,
			question, slug, outcomes, start_date, end_date,
			gamma_market_id, enable_order_book, accepting_orders, ready, funded,
			order_min_size, order_price_min_tick_size,
			best_bid, best_ask, last_trade_price, volume_clob, volume_num,
			status, winning_outcome, price_to_beat, last_book_hash,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22,
			$23, $24, $25, $26,
			now()
		) ON CONFLICT (slug) DO UPDATE SET
			condition_id = COALESCE(EXCLUDED.condition_id, new_markets.condition_id),
			token_id_yes = COALESCE(EXCLUDED.token_id_yes, new_markets.token_id_yes),
			token_id_no = COALESCE(EXCLUDED.token_id_no, new_markets.token_id_no),
			question = COALESCE(EXCLUDED.question, new_markets.question),
			outcomes = COALESCE(EXCLUDED.outcomes, new_markets.outcomes),
			start_date = COALESCE(EXCLUDED.start_date, new_markets.start_date),
			end_date = COALESCE(EXCLUDED.end_date, new_markets.end_date),
			gamma_market_id = COALESCE(EXCLUDED.gamma_market_id, new_markets.gamma_market_id),
			enable_order_book = COALESCE(EXCLUDED.enable_order_book, new_markets.enable_order_book),
			accepting_orders = COALESCE(EXCLUDED.accepting_orders, new_markets.accepting_orders),
			ready = COALESCE(EXCLUDED.ready, new_markets.ready),
			funded = COALESCE(EXCLUDED.funded, new_markets.funded),
			order_min_size = COALESCE(EXCLUDED.order_min_size, new_markets.order_min_size),
			order_price_min_tick_size = COALESCE(EXCLUDED.order_price_min_tick_size, new_markets.order_price_min_tick_size),
			best_bid = COALESCE(EXCLUDED.best_bid, new_markets.best_bid),
			best_ask = COALESCE(EXCLUDED.best_ask, new_markets.best_ask),
			last_trade_price = COALESCE(EXCLUDED.last_trade_price, new_markets.last_trade_price),
			volume_clob = COALESCE(EXCLUDED.volume_clob, new_markets.volume_clob),
			volume_num = COALESCE(EXCLUDED.volume_num, new_markets.volume_num),
			status = COALESCE(EXCLUDED.status, new_markets.status),
			winning_outcome = COALESCE(EXCLUDED.winning_outcome, new_markets.winning_outcome),
			price_to_beat = COALESCE(EXCLUDED.price_to_beat, new_markets.price_to_beat),
			last_book_hash = COALESCE(EXCLUDED.last_book_hash, new_markets.last_book_hash),
			updated_at = EXCLUDED.updated_at
	`
	_, err := tx.Exec(ctx, query,
		m.Symbol, m.Interval, m.ConditionID, m.TokenIDYes, m.TokenIDNo,
		m.Question, m.Slug, m.Outcomes, m.StartDate, m.EndDate,
		m.GammaMarketID, m.EnableOrderBook, m.AcceptingOrders, m.Ready, m.Funded,
		m.OrderMinSize, m.OrderPriceMinTickSize,
		m.BestBid, m.BestAsk, m.LastTradePrice, m.VolumeCLOB, m.VolumeNum,
		m.Status, m.WinningOutcome, m.PriceToBeat, m.LastBookHash,
	)
	if err != nil {
		return fmt.Errorf("upsert by slug tx: %w", err)
	}
	return nil
}

// GetMarketsNeedingRefresh returns markets that need their data refreshed.
func (s *PostgresStore) GetMarketsNeedingRefresh(ctx context.Context, maxAge time.Duration, limit int) ([]*models.NewMarket, error) {
	cutoff := time.Now().UTC().Add(-maxAge)

	query := `
		SELECT id, symbol, interval, condition_id, token_id_yes, token_id_no,
			question, slug, outcomes, start_date, end_date,
			gamma_market_id, enable_order_book, accepting_orders, ready, funded,
			order_min_size, order_price_min_tick_size,
			best_bid, best_ask, last_trade_price, volume_clob, volume_num,
			status, winning_outcome, price_to_beat, last_book_hash,
			created_at, updated_at
		FROM public.new_markets
		WHERE status = 'upcoming' OR updated_at < $1
		ORDER BY updated_at ASC
		LIMIT $2
	`

	rows, err := s.pool.Query(ctx, query, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("querying markets to refresh: %w", err)
	}
	defer rows.Close()

	return scanMarkets(rows)
}

// GetMarketsByTokenIDs returns markets for the given token_id_yes values.
func (s *PostgresStore) GetMarketsByTokenIDs(ctx context.Context, tokenIDs []string) (map[string]*models.NewMarket, error) {
	if len(tokenIDs) == 0 {
		return map[string]*models.NewMarket{}, nil
	}

	query := `
		SELECT id, symbol, interval, condition_id, token_id_yes, token_id_no,
			question, slug, outcomes, start_date, end_date,
			gamma_market_id, enable_order_book, accepting_orders, ready, funded,
			order_min_size, order_price_min_tick_size,
			best_bid, best_ask, last_trade_price, volume_clob, volume_num,
			status, winning_outcome, price_to_beat, last_book_hash,
			created_at, updated_at
		FROM public.new_markets
		WHERE token_id_yes = ANY($1)
	`

	rows, err := s.pool.Query(ctx, query, tokenIDs)
	if err != nil {
		return nil, fmt.Errorf("querying markets by token IDs: %w", err)
	}
	defer rows.Close()

	markets, err := scanMarkets(rows)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.NewMarket, len(markets))
	for _, m := range markets {
		if m.TokenIDYes != nil {
			result[*m.TokenIDYes] = m
		}
	}
	return result, nil
}

func scanMarkets(rows pgx.Rows) ([]*models.NewMarket, error) {
	var markets []*models.NewMarket
	for rows.Next() {
		m := &models.NewMarket{}
		err := rows.Scan(
			&m.ID, &m.Symbol, &m.Interval, &m.ConditionID, &m.TokenIDYes, &m.TokenIDNo,
			&m.Question, &m.Slug, &m.Outcomes, &m.StartDate, &m.EndDate,
			&m.GammaMarketID, &m.EnableOrderBook, &m.AcceptingOrders, &m.Ready, &m.Funded,
			&m.OrderMinSize, &m.OrderPriceMinTickSize,
			&m.BestBid, &m.BestAsk, &m.LastTradePrice, &m.VolumeCLOB, &m.VolumeNum,
			&m.Status, &m.WinningOutcome, &m.PriceToBeat, &m.LastBookHash,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning market row: %w", err)
		}
		markets = append(markets, m)
	}
	return markets, rows.Err()
}

// GetActiveMarketTokens returns all active token IDs (yes + no) from active markets.
// Returns 56 rows: 28 yes + 28 no.
// Coins and intervals are hardcoded — add/remove manually if needed.
func (s *PostgresStore) GetActiveMarketTokens(ctx context.Context) ([]models.ActiveToken, error) {
	query := `
		SELECT
			token_id_yes AS token_id, 'yes' AS side, symbol, interval
		FROM new_markets
		WHERE symbol IN ('btc','eth','sol','xrp','doge','hype','bnb')
		  AND interval IN ('5m','15m','1h','4h')
		  AND token_id_yes IS NOT NULL
		  AND status NOT IN ('closed', 'resolved')
		  AND enable_order_book = true
		UNION ALL
		SELECT
			token_id_no AS token_id, 'no' AS side, symbol, interval
		FROM new_markets
		WHERE symbol IN ('btc','eth','sol','xrp','doge','hype','bnb')
		  AND interval IN ('5m','15m','1h','4h')
		  AND token_id_no IS NOT NULL
		  AND status NOT IN ('closed', 'resolved')
		  AND enable_order_book = true
		ORDER BY symbol, interval, side
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying active market tokens: %w", err)
	}
	defer rows.Close()

	var tokens []models.ActiveToken
	for rows.Next() {
		var tok models.ActiveToken
		if err := rows.Scan(&tok.TokenID, &tok.Side, &tok.Symbol, &tok.Interval); err != nil {
			return nil, fmt.Errorf("scanning active token row: %w", err)
		}
		tokens = append(tokens, tok)
	}
	return tokens, rows.Err()
}

// InsertBookSnapshot inserts a full order book snapshot into book_snapshots.
func (s *PostgresStore) InsertBookSnapshot(ctx context.Context, snap *models.BookSnapshot) error {
	query := `
		INSERT INTO book_snapshots
			(token_id, side, symbol, interval, best_bid, best_ask, spread,
			 bid_size, ask_size, last_trade, book_hash, raw_bids, raw_asks)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := s.pool.Exec(ctx, query,
		snap.TokenID, snap.Side, snap.Symbol, snap.Interval,
		snap.BestBid, snap.BestAsk, snap.Spread,
		snap.BidSize, snap.AskSize, snap.LastTrade,
		snap.BookHash, snap.RawBids, snap.RawAsks,
	)
	if err != nil {
		return fmt.Errorf("inserting book snapshot: %w", err)
	}
	return nil
}

// UpsertCurrentBook upserts into current_books (ON CONFLICT on token_id).
func (s *PostgresStore) UpsertCurrentBook(ctx context.Context, snap *models.BookSnapshot) error {
	query := `
		INSERT INTO current_books
			(token_id, side, symbol, interval, best_bid, best_ask, spread,
			 bid_size, ask_size, last_trade, book_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (token_id) DO UPDATE SET
			side = EXCLUDED.side, symbol = EXCLUDED.symbol, interval = EXCLUDED.interval,
			best_bid = EXCLUDED.best_bid, best_ask = EXCLUDED.best_ask,
			spread = EXCLUDED.spread, bid_size = EXCLUDED.bid_size,
			ask_size = EXCLUDED.ask_size, last_trade = EXCLUDED.last_trade,
			book_hash = EXCLUDED.book_hash, updated_at = now()
	`
	_, err := s.pool.Exec(ctx, query,
		snap.TokenID, snap.Side, snap.Symbol, snap.Interval,
		snap.BestBid, snap.BestAsk, snap.Spread,
		snap.BidSize, snap.AskSize, snap.LastTrade, snap.BookHash,
	)
	if err != nil {
		return fmt.Errorf("upserting current book: %w", err)
	}
	return nil
}

// InsertBookSnapshotsBatch inserts multiple book snapshots using a single pgx.Batch.
func (s *PostgresStore) InsertBookSnapshotsBatch(ctx context.Context, snaps []*models.BookSnapshot) error {
	if len(snaps) == 0 {
		return nil
	}

	const query = `
		INSERT INTO book_snapshots
			(token_id, side, symbol, interval, best_bid, best_ask, spread,
			 bid_size, ask_size, last_trade, book_hash, raw_bids, raw_asks)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	batch := &pgx.Batch{}
	for _, snap := range snaps {
		batch.Queue(query,
			snap.TokenID, snap.Side, snap.Symbol, snap.Interval,
			snap.BestBid, snap.BestAsk, snap.Spread,
			snap.BidSize, snap.AskSize, snap.LastTrade,
			snap.BookHash, snap.RawBids, snap.RawAsks,
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	// Drain results
	for i := 0; i < len(snaps); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert book_snapshots (row %d): %w", i, err)
		}
	}
	return nil
}

// UpsertCurrentBooksBatch upserts multiple current_books rows using a single pgx.Batch.
func (s *PostgresStore) UpsertCurrentBooksBatch(ctx context.Context, snaps []*models.BookSnapshot) error {
	if len(snaps) == 0 {
		return nil
	}

	const query = `
		INSERT INTO current_books
			(token_id, side, symbol, interval, best_bid, best_ask, spread,
			 bid_size, ask_size, last_trade, book_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (token_id) DO UPDATE SET
			side = EXCLUDED.side, symbol = EXCLUDED.symbol, interval = EXCLUDED.interval,
			best_bid = EXCLUDED.best_bid, best_ask = EXCLUDED.best_ask,
			spread = EXCLUDED.spread, bid_size = EXCLUDED.bid_size,
			ask_size = EXCLUDED.ask_size, last_trade = EXCLUDED.last_trade,
			book_hash = EXCLUDED.book_hash, updated_at = now()
	`

	batch := &pgx.Batch{}
	for _, snap := range snaps {
		batch.Queue(query,
			snap.TokenID, snap.Side, snap.Symbol, snap.Interval,
			snap.BestBid, snap.BestAsk, snap.Spread,
			snap.BidSize, snap.AskSize, snap.LastTrade, snap.BookHash,
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(snaps); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch upsert current_books (row %d): %w", i, err)
		}
	}
	return nil
}
