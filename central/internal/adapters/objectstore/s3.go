// Package objectstore is the S3-compatible ObjectWriter adapter
// (OpenSpec add-central-go-platform, task 4.3). Keys are derived from
// the content hash, so every object is immutable by construction; the
// upload is verified by reading it back, and a failed verification
// cleans the object up before reporting the error.
package objectstore

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/ports"
)

// S3Writer talks plain path-style HTTP to any S3-compatible endpoint.
type S3Writer struct {
	Endpoint string // e.g. http://127.0.0.1:9000
	Bucket   string
	Client   *http.Client
}

var _ ports.ObjectWriter = S3Writer{}

// WriteSnapshot encodes, compresses, uploads and VERIFIES the
// canonical snapshot, returning its content-addressed reference.
func (w S3Writer) WriteSnapshot(ctx context.Context,
	snap catalog.Snapshot) (catalog.ObjectRef, error) {
	body, sum, err := encode(snap)
	if err != nil {
		return catalog.ObjectRef{}, err
	}
	key := fmt.Sprintf("catalog/%s.json.gz", sum)
	if err := w.put(ctx, key, body, sum); err != nil {
		return catalog.ObjectRef{}, err
	}
	if err := w.verify(ctx, key, sum); err != nil {
		w.cleanup(ctx, key)
		return catalog.ObjectRef{}, err
	}
	return catalog.ObjectRef{Key: key, SHA256: sum,
		Size: int64(len(body))}, nil
}

// encode renders the canonical JSON and compresses it; the hash of
// the compressed bytes IS the object identity.
func encode(snap catalog.Snapshot) ([]byte, string, error) {
	raw, err := json.Marshal(snap)
	if err != nil {
		return nil, "", fmt.Errorf("objectstore: encode: %w", err)
	}
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, "", fmt.Errorf("objectstore: gzip: %w", err)
	}
	if _, err := zw.Write(raw); err != nil {
		return nil, "", fmt.Errorf("objectstore: compress: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, "", fmt.Errorf("objectstore: flush: %w", err)
	}
	sum := sha256.Sum256(buf.Bytes())
	return buf.Bytes(), hex.EncodeToString(sum[:]), nil
}

func (w S3Writer) put(ctx context.Context, key string, body []byte,
	sum string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		w.url(key), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("objectstore: put request: %w", err)
	}
	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("x-amz-content-sha256", sum)
	resp, err := w.client().Do(req)
	if err != nil {
		return fmt.Errorf("objectstore: put: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("objectstore: put %s: status %d", key, resp.StatusCode)
	}
	return nil
}
