package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/shahabbasian/polymarket-market-fetcher/internal/api"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/collector"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/config"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/db"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/logging"
	"github.com/shahabbasian/polymarket-market-fetcher/internal/scanner"
)

func main() {
	cfg := config.Load()
	logger := logging.Setup(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(logger)

	slog.Info("starting polymarket market fetcher",
		"gamma_base_url", cfg.GammaBaseURL,
		"clob_base_url", cfg.CLOBBaseURL,
		"discovery_interval", cfg.DiscoveryInterval,
		"refresh_interval", cfg.RefreshInterval,
	)

	// Setup early signal channel for interrupt during DB connect
	earlySigCh := make(chan os.Signal, 1)
	signal.Notify(earlySigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(earlySigCh)

	// Connect to PostgreSQL with infinite retry
	var store *db.PostgresStore
	for {
		var err error
		store, err = db.NewPostgresStore(context.Background(), cfg.DatabaseURL)
		if err == nil {
			slog.Info("connected to database")
			break
		}
		slog.Error("failed to connect to database, retrying", "error", err, "retry_in", "5s")
		select {
		case sig := <-earlySigCh:
			slog.Info("received signal during startup", "signal", sig.String())
			os.Exit(1)
		case <-time.After(5 * time.Second):
		}
	}
	defer store.Close()

	// Migrations are intentionally disabled in this project.
	// Database schema changes are performed manually by the operator.
	slog.Info("migrations skipped: manual schema management only")

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: cfg.HTTPTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Create API clients with rate limiters
	gammaRL := api.NewRateLimiter(cfg.RateLimitDelay, cfg.BackoffMaxDelay)
	clobRL := api.NewRateLimiter(cfg.RateLimitDelay, cfg.BackoffMaxDelay)
	gammaClient := api.NewGammaClient(httpClient, cfg.GammaBaseURL, gammaRL)
	clobClient := api.NewCLOBClient(httpClient, cfg.CLOBBaseURL, clobRL)

	// Separate CLOB client for collector with its own rate limiter (3 req/s)
	collectorClobRL := api.NewRateLimiter(time.Duration(cfg.CollectorDelayMS)*time.Millisecond, cfg.BackoffMaxDelay)
	collectorClobClient := api.NewCLOBClient(httpClient, cfg.CLOBBaseURL, collectorClobRL)

	// Create and start scanner
	sc := scanner.New(store, gammaClient, clobClient, cfg)

	// Create collector with its own rate limiter (333ms) and CLOB client
	bookCollector := collector.New(store, collectorClobClient, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sc.Run(ctx)
	bookCollector.Run(ctx)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	slog.Info("received shutdown signal", "signal", sig.String())
	cancel()

	// Wait for scanner and collector to finish current work
	done := make(chan struct{})
	go func() {
		sc.Wait()
		bookCollector.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("graceful shutdown complete")
	case <-time.After(30 * time.Second):
		slog.Warn("shutdown timeout exceeded, exiting")
	}
}
