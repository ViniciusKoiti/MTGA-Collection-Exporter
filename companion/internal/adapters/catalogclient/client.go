package catalogclient

import (
	"context"
	"errors"
	"fmt"
)

// Fetcher downloads the manifest and its object; the harness wires a
// controlled one, production wires HTTPS.
type Fetcher interface {
	Manifest(ctx context.Context) (Manifest, error)
	Object(ctx context.Context, key string) ([]byte, error)
}

// Cache persists the last VERIFIED catalog; nothing unverified is
// ever stored.
type Cache interface {
	Store(man Manifest, object []byte) error
	Load() (Manifest, []byte, bool)
}

// Client refreshes the catalog, verifying before caching. When the
// network — or the verification — fails, it falls back to the cached
// verified catalog and says so via Stale.
type Client struct {
	Fetch  Fetcher
	Cache  Cache
	Verify Verifier
	Schema string
}

// Result is the catalog the desktop should use right now.
type Result struct {
	Manifest Manifest
	Object   []byte
	Stale    bool
}

// Current returns a fresh verified catalog, or the cached one marked
// stale, or an error when neither exists.
func (c Client) Current(ctx context.Context) (Result, error) {
	fresh, err := c.refresh(ctx)
	if err == nil {
		return fresh, nil
	}
	man, object, ok := c.Cache.Load()
	if !ok {
		return Result{}, fmt.Errorf(
			"catalogclient: no verified catalog available: %w", err)
	}
	return Result{Manifest: man, Object: object, Stale: true}, nil
}

func (c Client) refresh(ctx context.Context) (Result, error) {
	man, err := c.Fetch.Manifest(ctx)
	if err != nil {
		return Result{}, err
	}
	object, err := c.Fetch.Object(ctx, man.Object.Key)
	if err != nil {
		return Result{}, err
	}
	if err := c.Verify.Verify(man, object, c.Schema); err != nil {
		return Result{}, err
	}
	if err := c.Cache.Store(man, object); err != nil {
		return Result{}, errors.Join(
			errors.New("catalogclient: cache store failed"), err)
	}
	return Result{Manifest: man, Object: object}, nil
}
