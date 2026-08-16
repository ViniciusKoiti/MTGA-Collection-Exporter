package scryfall

import (
	"context"
	"fmt"
	"time"
)

// OpenCachedOnly serves whatever cache exists WITHOUT touching the
// network: the deck workspace uses it so a click can never trigger a
// bulk download. Freshness is unknown without a refresh, so the
// catalog is flagged Stale; no cache at all is an error.
func OpenCachedOnly(dir string) (*Catalog, error) {
	catalog, err := loadCache(dir)
	if err != nil {
		return nil, err
	}
	catalog.Stale = true
	return catalog, nil
}

// Open is the composition entry point: refresh the cache when it is
// missing or older than maxAge, and fall back to the last good cache —
// flagged Stale — when the network is unavailable. No cache and no
// network is a hard error; the catalog never invents data.
func Open(
	ctx context.Context,
	client *Client,
	now time.Time,
	maxAge time.Duration,
) (*Catalog, error) {
	cached, cacheErr := loadCache(client.Dir)
	if cacheErr == nil && cached.Meta.Fresh(now, maxAge) {
		return cached, nil
	}
	if _, err := client.Refresh(ctx, now); err != nil {
		if cacheErr == nil {
			cached.Stale = true // offline fallback: serve the last good data
			return cached, nil
		}
		return nil, fmt.Errorf("scryfall: no usable catalog (network: %w)", err)
	}
	refreshed, err := loadCache(client.Dir)
	if err != nil {
		return nil, err
	}
	return refreshed, nil
}
