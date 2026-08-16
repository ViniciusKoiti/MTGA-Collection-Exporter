package objectstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

func (w S3Writer) url(key string) string {
	return fmt.Sprintf("%s/%s/%s", w.Endpoint, w.Bucket, key)
}

func (w S3Writer) client() *http.Client {
	if w.Client != nil {
		return w.Client
	}
	return http.DefaultClient
}

// verify reads the object back and compares its hash with the one we
// uploaded — an upload is only real once its bytes prove themselves.
func (w S3Writer) verify(ctx context.Context, key, wantSum string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		w.url(key), nil)
	if err != nil {
		return fmt.Errorf("objectstore: verify request: %w", err)
	}
	resp, err := w.client().Do(req)
	if err != nil {
		return fmt.Errorf("objectstore: verify: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("objectstore: verify %s: status %d",
			key, resp.StatusCode)
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, resp.Body); err != nil {
		return fmt.Errorf("objectstore: verify read: %w", err)
	}
	if got := hex.EncodeToString(hasher.Sum(nil)); got != wantSum {
		return fmt.Errorf("objectstore: verification mismatch for %s", key)
	}
	return nil
}

// cleanup deletes a failed upload on a best-effort basis; the error
// path that triggered it is the one reported to the caller.
func (w S3Writer) cleanup(ctx context.Context, key string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		w.url(key), nil)
	if err != nil {
		return
	}
	if resp, err := w.client().Do(req); err == nil {
		resp.Body.Close()
	}
}
