package scanner

import (
	"fmt"
	"strconv"
	"time"
)

// GenerateInitialSlugs creates N future slugs anchored to current time.
func GenerateInitialSlugs(symbol string, interval string, n int) ([]string, error) {
	var dur int64
	switch interval {
	case "5m":
		dur = 300
	case "15m":
		dur = 900
	default:
		return nil, fmt.Errorf("unknown interval: %s", interval)
	}

	now := time.Now().UTC()
	ts := now.Unix()
	rounded := (ts / dur) * dur

	var slugs []string
	for i := int64(1); i <= int64(n); i++ {
		nextTs := rounded + i*dur
		slug := fmt.Sprintf("%s-updown-%s-%d", symbol, interval, nextTs)
		slugs = append(slugs, slug)
	}
	return slugs, nil
}

// GenerateNextSlugs generates N slugs starting from a known slug.
func GenerateNextSlugs(symbol string, interval string, currentSlug string, n int) ([]string, error) {
	var dur int64
	switch interval {
	case "5m":
		dur = 300
	case "15m":
		dur = 900
	default:
		return nil, fmt.Errorf("unknown interval: %s", interval)
	}

	ts, err := parseTimestampSlug(currentSlug, interval)
	if err != nil {
		return nil, err
	}

	// If anchor is so old that even N future slugs are expired, it's stale.
	anchorTime := time.Unix(ts, 0)
	staleThreshold := time.Now().UTC().Add(-time.Duration(n) * time.Duration(dur) * time.Second)
	if anchorTime.Before(staleThreshold) {
		return nil, fmt.Errorf("slug anchor %s is stale (older than %d %s intervals)", currentSlug, n, interval)
	}

	var slugs []string
	for i := int64(1); i <= int64(n); i++ {
		nextTs := ts + i*dur
		slug := fmt.Sprintf("%s-updown-%s-%d", symbol, interval, nextTs)
		slugs = append(slugs, slug)
	}
	return slugs, nil
}

// GenerateNextSlug returns the single slug immediately after currentSlug.
func GenerateNextSlug(symbol string, interval string, currentSlug string) (string, error) {
	slugs, err := GenerateNextSlugs(symbol, interval, currentSlug, 1)
	if err != nil {
		return "", err
	}
	if len(slugs) == 0 {
		return "", fmt.Errorf("no slug generated")
	}
	return slugs[0], nil
}

func parseTimestampSlug(slug string, interval string) (int64, error) {
	prefix := fmt.Sprintf("-updown-%s-", interval)
	idx := len(prefix)
	if len(slug) < idx+1 {
		return 0, fmt.Errorf("slug too short: %s", slug)
	}

	// Find the last occurrence of the prefix
	for i := len(slug) - len(prefix) - 1; i >= 0; i-- {
		if slug[i:i+len(prefix)] == prefix {
			tsStr := slug[i+len(prefix):]
			ts, err := strconv.ParseInt(tsStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid timestamp in slug %s: %w", slug, err)
			}
			return ts, nil
		}
	}

	return 0, fmt.Errorf("no timestamp found in slug: %s", slug)
}

// CurrentSlug returns the slug for the currently active market window.
// Returns empty string if symbol or interval is invalid.
func CurrentSlug(symbol string, interval string) string {
	var dur int64
	switch interval {
	case "5m":
		dur = 300
	case "15m":
		dur = 900
	default:
		return ""
	}

	now := time.Now().UTC()
	ts := now.Unix()
	rounded := (ts / dur) * dur

	return fmt.Sprintf("%s-updown-%s-%d", symbol, interval, rounded)
}
