package scanner

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/shahabbasian/polymarket-market-fetcher/internal/models"
)

// GenerateInitialSlugs creates N future slugs anchored to current time.
func GenerateInitialSlugs(symbol string, interval string, n int) ([]string, error) {
	now := time.Now().UTC()

	if interval == "1h" {
		return generateInitial1hSlugs(symbol, now, n)
	}

	var dur int64
	switch interval {
	case "5m":
		dur = 300
	case "15m":
		dur = 900
	case "4h":
		dur = 14400
	default:
		return nil, fmt.Errorf("unknown interval: %s", interval)
	}

	// Anchor to nearest past multiple of duration
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
	if interval == "1h" {
		return generateNext1hSlugs(symbol, currentSlug, n)
	}

	var dur int64
	switch interval {
	case "5m":
		dur = 300
	case "15m":
		dur = 900
	case "4h":
		dur = 14400
	default:
		return nil, fmt.Errorf("unknown interval: %s", interval)
	}

	ts, err := parseTimestampSlug(currentSlug, interval)
	if err != nil {
		return nil, err
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

// ---- 1h (human-readable date) patterns ----

var hour1hRegex = regexp.MustCompile(`^(\w+)-up-or-down-([a-z]{3})-(\d{1,2})-(\d{4})-(\d{1,2})([ap]m)-et$`)

// Month abbreviation to number mapping.
var monthMap = map[string]time.Month{
	"jan": time.January,
	"feb": time.February,
	"mar": time.March,
	"apr": time.April,
	"may": time.May,
	"jun": time.June,
	"jul": time.July,
	"aug": time.August,
	"sep": time.September,
	"oct": time.October,
	"nov": time.November,
	"dec": time.December,
}

var monthAbbrs = []string{"", "jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"}

func generateInitial1hSlugs(symbol string, now time.Time, n int) ([]string, error) {
	fullName := models.CoinSymbolMap[symbol]
	if fullName == "" {
		return nil, fmt.Errorf("unknown symbol: %s", symbol)
	}

	// Anchor to the current hour
	anchor := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)

	var slugs []string
	for i := 1; i <= n; i++ {
		t := anchor.Add(time.Duration(i) * time.Hour)
		slug := format1hSlug(fullName, t)
		slugs = append(slugs, slug)
	}
	return slugs, nil
}

func generateNext1hSlugs(symbol string, currentSlug string, n int) ([]string, error) {
	fullName := models.CoinSymbolMap[symbol]
	if fullName == "" {
		return nil, fmt.Errorf("unknown symbol: %s", symbol)
	}

	// Try to parse from existing slug
	t, err := parse1hSlug(currentSlug)
	if err != nil {
		// Fallback: use current time and just generate from now
		return generateInitial1hSlugs(symbol, time.Now().UTC(), n)
	}

	var slugs []string
	for i := 1; i <= n; i++ {
		next := t.Add(time.Duration(i) * time.Hour)
		slug := format1hSlug(fullName, next)
		slugs = append(slugs, slug)
	}
	return slugs, nil
}

func parse1hSlug(slug string) (time.Time, error) {
	matches := hour1hRegex.FindStringSubmatch(slug)
	if len(matches) != 7 {
		return time.Time{}, fmt.Errorf("slug does not match 1h pattern: %s", slug)
	}

	// matches[1] = fullName (not used here)
	monthAbbr := matches[2]
	dayStr := matches[3]
	yearStr := matches[4]
	hourStr := matches[5]
	ampm := matches[6]

	month, ok := monthMap[monthAbbr]
	if !ok {
		return time.Time{}, fmt.Errorf("unknown month abbreviation: %s", monthAbbr)
	}

	day, _ := strconv.Atoi(dayStr)
	year, _ := strconv.Atoi(yearStr)
	hour, _ := strconv.Atoi(hourStr)

	if ampm == "pm" && hour != 12 {
		hour += 12
	} else if ampm == "am" && hour == 12 {
		hour = 0
	}

	t := time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	return t, nil
}

func format1hSlug(fullName string, t time.Time) string {
	hour := t.Hour()
	var hourStr string
	if hour == 0 {
		hourStr = "12am"
	} else if hour < 12 {
		hourStr = fmt.Sprintf("%dam", hour)
	} else if hour == 12 {
		hourStr = "12pm"
	} else {
		hourStr = fmt.Sprintf("%dpm", hour-12)
	}

	return fmt.Sprintf("%s-up-or-down-%s-%d-%d-%s-et",
		fullName,
		monthAbbrs[t.Month()],
		t.Day(),
		t.Year(),
		hourStr,
	)
}
