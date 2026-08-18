package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Document is one public payload (manifest or compatibility matrix).
type Document struct {
	Body        []byte
	ContentType string
}

// DocumentSource produces the current document; it runs only on cache
// miss and concurrent misses collapse into one call via singleflight.
type DocumentSource func(ctx context.Context) (Document, error)

// DocHandler serves one public document with ETag revalidation and a
// single-entry cache — bounded by construction, no key space to grow.
type DocHandler struct {
	Source    DocumentSource
	TTL       time.Duration
	MaxAge    int
	Immutable bool

	mu      sync.Mutex
	cached  Document
	etag    string
	expires time.Time
	group   singleflight.Group
}

type cachedDoc struct {
	doc  Document
	etag string
}

func (h *DocHandler) current(ctx context.Context) (cachedDoc, error) {
	h.mu.Lock()
	if h.etag != "" && time.Now().Before(h.expires) {
		out := cachedDoc{doc: h.cached, etag: h.etag}
		h.mu.Unlock()
		return out, nil
	}
	h.mu.Unlock()
	v, err, _ := h.group.Do("doc", func() (any, error) {
		doc, err := h.Source(ctx)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(doc.Body)
		out := cachedDoc{doc: doc,
			etag: fmt.Sprintf("%q", hex.EncodeToString(sum[:16]))}
		h.mu.Lock()
		h.cached, h.etag, h.expires = doc, out.etag, time.Now().Add(h.TTL)
		h.mu.Unlock()
		return out, nil
	})
	if err != nil {
		return cachedDoc{}, err
	}
	return v.(cachedDoc), nil
}

func (h *DocHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		WriteError(w, r, http.StatusMethodNotAllowed,
			"method_not_allowed", "read-only endpoint")
		return
	}
	cur, err := h.current(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusServiceUnavailable,
			"unavailable", "document source unavailable")
		return
	}
	cacheControl := fmt.Sprintf("public, max-age=%d", h.MaxAge)
	if h.Immutable {
		cacheControl += ", immutable"
	}
	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("ETag", cur.etag)
	if r.Header.Get("If-None-Match") == cur.etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", cur.doc.ContentType)
	_, _ = w.Write(cur.doc.Body)
}
