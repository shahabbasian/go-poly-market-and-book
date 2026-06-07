package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL        string
	GammaBaseURL       string
	CLOBBaseURL        string
	RateLimitDelay     time.Duration
	DiscoveryInterval  time.Duration
	RefreshInterval    time.Duration
	FutureSlugCount    int
	InitialSlugCount   int
	CLOBBatchSize      int
	RefreshMaxAge      time.Duration
	RefreshLimit       int
	LogLevel           string
	LogFormat          string
	HTTPTimeout        time.Duration
	MaxRetries         int
	BackoffBaseDelay   time.Duration
	BackoffMaxDelay    time.Duration
	CollectorDelayMS       int
	CollectorFlushIntervalS int
	CollectorTokenRefreshS  int
	CollectorBufferMaxRows  int
}

func Load() *Config {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL environment variable is required")
	}

	return &Config{
		DatabaseURL:       databaseURL,
		GammaBaseURL:      envString("GAMMA_BASE_URL", "https://gamma-api.polymarket.com"),
		CLOBBaseURL:       envString("CLOB_BASE_URL", "https://clob.polymarket.com"),
		RateLimitDelay:    envDurationMS("RATE_LIMIT_DELAY_MS", 250),
		DiscoveryInterval: envDurationS("DISCOVERY_INTERVAL_S", 30),
		RefreshInterval:   envDurationS("REFRESH_INTERVAL_S", 60),
		FutureSlugCount:   envInt("FUTURE_SLUG_COUNT", 10),
		InitialSlugCount:  envInt("INITIAL_SLUG_COUNT", 20),
		CLOBBatchSize:     envInt("CLOB_BATCH_SIZE", 50),
		RefreshMaxAge:     envDurationM("REFRESH_MAX_AGE_M", 5),
		RefreshLimit:      envInt("REFRESH_LIMIT", 100),
		LogLevel:          envString("LOG_LEVEL", "info"),
		LogFormat:         envString("LOG_FORMAT", "json"),
		HTTPTimeout:       envDurationS("HTTP_TIMEOUT_S", 30),
		MaxRetries:        envInt("MAX_RETRIES", 5),
		BackoffBaseDelay:  envDurationMS("BACKOFF_BASE_DELAY_MS", 1000),
		BackoffMaxDelay:   envDurationMS("BACKOFF_MAX_DELAY_MS", 30000),
		// Backtest mode: aggressive polling for high-resolution order book data.
		// 10 Hz polling per market-side captures wicks, slippage, and order flow.
		// Monitor logs for 429 rate-limit warnings; raise to 200ms if needed.
		CollectorDelayMS:       envInt("COLLECTOR_DELAY_MS", 100),
		CollectorFlushIntervalS: envInt("COLLECTOR_FLUSH_INTERVAL_S", 2),
		CollectorTokenRefreshS:  envInt("COLLECTOR_TOKEN_REFRESH_S", 60),
		CollectorBufferMaxRows:  envInt("COLLECTOR_BUFFER_MAX_ROWS", 5000),
	}
}

func envString(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func envDurationMS(key string, defaultMS int) time.Duration {
	return time.Duration(envInt(key, defaultMS)) * time.Millisecond
}

func envDurationS(key string, defaultS int) time.Duration {
	return time.Duration(envInt(key, defaultS)) * time.Second
}

func envDurationM(key string, defaultM int) time.Duration {
	return time.Duration(envInt(key, defaultM)) * time.Minute
}
