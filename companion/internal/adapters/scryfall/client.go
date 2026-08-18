package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Metadata records the provenance and freshness of the cached bulk data.
type Metadata struct {
	Version   string    `json:"version"`    // bulk updated_at from Scryfall
	FetchedAt time.Time `json:"fetched_at"` // local fetch instant
}

// Fresh reports whether the cache is inside the freshness window.
func (m Metadata) Fresh(now time.Time, maxAge time.Duration) bool {
	return !m.FetchedAt.IsZero() && now.Sub(m.FetchedAt) <= maxAge
}

// Client downloads the default_cards bulk file with a bounded retry
// budget; the HTTP client (and its timeout) is injected by composition.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	Dir     string
	Retries int
}

// Refresh fetches the bulk index and card file, then atomically replaces
// the local cache. Every attempt honors ctx; the retry budget is hard.
func (c *Client) Refresh(ctx context.Context, now time.Time) (Metadata, error) {
	var lastErr error
	attempts := max(c.Retries, 1)
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return Metadata{}, err
		}
		meta, err := c.fetchOnce(ctx, now)
		if err == nil {
			return meta, nil
		}
		lastErr = err
	}
	return Metadata{}, fmt.Errorf("scryfall: retry budget exhausted: %w", lastErr)
}

func (c *Client) fetchOnce(ctx context.Context, now time.Time) (Metadata, error) {
	var index bulkIndex
	if err := c.getJSON(ctx, c.BaseURL+"/bulk-data", &index); err != nil {
		return Metadata{}, err
	}
	for _, entry := range index.Data {
		if entry.Type != "default_cards" {
			continue
		}
		var cards []bulkCard
		if err := c.getJSON(ctx, entry.DownloadURI, &cards); err != nil {
			return Metadata{}, err
		}
		meta := Metadata{Version: entry.UpdatedAt, FetchedAt: now}
		if err := storeCache(c.Dir, cards, meta); err != nil {
			return Metadata{}, err
		}
		return meta, nil
	}
	return Metadata{}, fmt.Errorf("scryfall: bulk index has no default_cards entry")
}

func (c *Client) getJSON(ctx context.Context, url string, out any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := c.HTTP.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("scryfall: %s returned %d", url, response.StatusCode)
	}
	payload, err := io.ReadAll(io.LimitReader(response.Body, 64<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, out)
}
