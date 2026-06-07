package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/shahabbasian/polymarket-market-fetcher/internal/models"
)

// CLOBClient fetches order book data from the CLOB API.
type CLOBClient struct {
	baseURL     string
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

func NewCLOBClient(httpClient *http.Client, baseURL string, rateLimiter *RateLimiter) *CLOBClient {
	return &CLOBClient{
		baseURL:     baseURL,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
	}
}

// FetchBooksBatch fetches order books for multiple token IDs in a single POST.
// Returns a map of token_id -> book response.
func (c *CLOBClient) FetchBooksBatch(ctx context.Context, tokenIDs []string) (map[string]*models.CLOBBookResponse, error) {
	if len(tokenIDs) == 0 {
		return map[string]*models.CLOBBookResponse{}, nil
	}

	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	reqBody := make([]models.CLOBBookRequest, 0, len(tokenIDs))
	for _, id := range tokenIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			reqBody = append(reqBody, models.CLOBBookRequest{TokenID: id})
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := c.baseURL + "/books"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			slog.Warn("clob rate limited (429)", "tokens", len(tokenIDs), "body", string(body))
		} else {
			slog.Warn("clob batch fetch failed", "status", resp.StatusCode, "body", string(body))
		}
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var results []models.CLOBBookResponse
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	m := make(map[string]*models.CLOBBookResponse, len(results))
	for i := range results {
		r := &results[i]
		m[r.AssetID] = r
	}

	return m, nil
}
