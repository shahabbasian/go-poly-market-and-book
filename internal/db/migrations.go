package db

import (
	"context"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations executes idempotent CREATE statements.
// Permission errors on index creation are logged as warnings and skipped
// so the app can still start if the DB user is not the table owner.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS public.new_markets (
			id uuid NOT NULL DEFAULT gen_random_uuid(),
			symbol character varying(10) NOT NULL,
			interval character varying(10) NOT NULL,
			condition_id character varying(128) NULL,
			token_id_yes character varying(128) NULL,
			token_id_no character varying(128) NULL,
			question text NULL,
			slug text NOT NULL,
			outcomes text NULL,
			start_date timestamp with time zone NULL,
			end_date timestamp with time zone NULL,
			gamma_market_id character varying(64) NULL,
			enable_order_book boolean NULL,
			accepting_orders boolean NULL,
			ready boolean NULL,
			funded boolean NULL,
			order_min_size double precision NULL,
			order_price_min_tick_size double precision NULL,
			best_bid double precision NULL,
			best_ask double precision NULL,
			last_trade_price double precision NULL,
			volume_clob double precision NULL,
			volume_num double precision NULL,
			created_at timestamp with time zone NOT NULL DEFAULT now(),
			updated_at timestamp with time zone NOT NULL DEFAULT now(),
			status character varying(20) NULL DEFAULT 'upcoming',
			winning_outcome character varying(10) NULL,
			price_to_beat double precision NULL,
			last_book_hash character varying(128) NULL,
			CONSTRAINT new_markets_pkey PRIMARY KEY (id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS unique_new_token_id_yes ON public.new_markets (token_id_yes) WHERE token_id_yes IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_new_markets_symbol_interval ON public.new_markets USING btree (symbol, interval)`,
		`CREATE INDEX IF NOT EXISTS idx_new_markets_start_date ON public.new_markets USING btree (start_date)`,
		`CREATE INDEX IF NOT EXISTS idx_new_markets_token_id_yes ON public.new_markets USING btree (token_id_yes)`,
		`CREATE INDEX IF NOT EXISTS idx_new_markets_status ON public.new_markets USING btree (status)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_new_markets_unique_slug ON public.new_markets (slug)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_new_markets_unique_symbol_interval_start ON public.new_markets (symbol, interval, start_date) WHERE start_date IS NOT NULL`,

		`CREATE TABLE IF NOT EXISTS public.book_snapshots (
			id          BIGSERIAL PRIMARY KEY,
			token_id    VARCHAR(128)  NOT NULL,
			side        VARCHAR(3)    NOT NULL,
			symbol      VARCHAR(10)   NOT NULL,
			interval    VARCHAR(10)   NOT NULL,
			best_bid    DOUBLE PRECISION,
			best_ask    DOUBLE PRECISION,
			spread      DOUBLE PRECISION,
			bid_size    DOUBLE PRECISION,
			ask_size    DOUBLE PRECISION,
			last_trade  DOUBLE PRECISION,
			book_hash   VARCHAR(128),
			raw_bids    JSONB,
			raw_asks    JSONB,
			recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_book_snapshots_token_time ON book_snapshots (token_id, recorded_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_book_snapshots_symbol_interval ON book_snapshots (symbol, interval, recorded_at DESC)`,

		`CREATE TABLE IF NOT EXISTS public.current_books (
			token_id    VARCHAR(128)  PRIMARY KEY,
			side        VARCHAR(3)    NOT NULL,
			symbol      VARCHAR(10)   NOT NULL,
			interval    VARCHAR(10)   NOT NULL,
			best_bid    DOUBLE PRECISION,
			best_ask    DOUBLE PRECISION,
			spread      DOUBLE PRECISION,
			bid_size    DOUBLE PRECISION,
			ask_size    DOUBLE PRECISION,
			last_trade  DOUBLE PRECISION,
			book_hash   VARCHAR(128),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
	}

	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			if isPermissionOrOwnershipError(err) {
				slog.Warn("Migration skipped: insufficient permissions (table may be owned by another user)",
					"stmt", firstLine(stmt),
					"error", err,
				)
				continue
			}
			slog.Error("Migration failed", "stmt", firstLine(stmt), "error", err)
			return err
		}
	}

	slog.Info("Migrations completed successfully")
	return nil
}

func isPermissionOrOwnershipError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "42501") || // insufficient_privilege
		strings.Contains(msg, "must be owner") ||
		strings.Contains(msg, "permission denied")
}

func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[:idx]
	}
	if len(s) > 100 {
		return s[:100] + "..."
	}
	return s
}
