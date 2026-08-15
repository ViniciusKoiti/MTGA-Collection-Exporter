package httpapi

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"

	"golang.org/x/sync/semaphore"
)

// withLimits enforces the concurrency, body and decompression budgets
// before the handler runs; every refusal is an explicit status code.
func withLimits(cfg Config, next http.Handler) http.Handler {
	slots := semaphore.NewWeighted(cfg.MaxConcurrent)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !slots.TryAcquire(1) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "server at capacity", http.StatusServiceUnavailable)
			return
		}
		defer slots.Release(1)
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodyBytes)
		if r.Header.Get("Content-Encoding") == "gzip" {
			body, err := cappedGzip(r.Body, cfg.MaxBodyBytes)
			if err != nil {
				http.Error(w, "invalid or oversized gzip body",
					http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = body
			r.Header.Del("Content-Encoding")
		}
		next.ServeHTTP(w, r)
	})
}

// errInflatedTooLarge marks a stream that inflates past the budget.
var errInflatedTooLarge = errors.New("httpapi: inflated body exceeds budget")

// cappedGzip inflates at most maxBytes of the request body; a stream
// inflating past the budget is refused (decompression bomb guard).
func cappedGzip(body io.ReadCloser, maxBytes int64) (io.ReadCloser, error) {
	defer body.Close()
	gz, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	inflated, err := io.ReadAll(io.LimitReader(gz, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(inflated)) > maxBytes {
		return nil, errInflatedTooLarge
	}
	return io.NopCloser(bytes.NewReader(inflated)), nil
}
