package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/shahabbasian/polymarket-market-fetcher/internal/models"
)

// NotFoundError indicates the market slug does not exist yet.
type NotFoundError struct{}

func (e NotFoundError) Error() string { return "market not found" }

// GammaClient fetches market data from the Gamma API.
type GammaClient struct {
	baseURL     string
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

func NewGammaClient(httpClient *http.Client, baseURL string, rateLimiter *RateLimiter) *GammaClient {
	return &GammaClient{
		baseURL:     baseURL,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
	}
}

// FetchMarketBySlug fetches a single market by slug.
// Returns (nil, nil) on 404 — market does not exist yet.
// Returns (nil, error) on transient errors (caller will retry).
func (c *GammaClient) FetchMarketBySlug(ctx context.Context, slug string) (*models.GammaMarketResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/markets/slug/%s", c.baseURL, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var m models.GammaMarketResponse
		if err := json.Unmarshal(body, &m); err != nil {
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		return &m, nil
	case http.StatusNotFound:
		return nil, NotFoundError{}
	case http.StatusTooManyRequests:
		slog.Warn("gamma rate limited", "slug", slug, "status", resp.StatusCode)
		return nil, fmt.Errorf("rate limited: %s", resp.Status)
	default:
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
}
