package httpapi

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RatePolicy is fixed-window admission control with bounded memory:
// requests beyond the per-principal limit — or beyond the key budget —
// are refused, never queued.
type RatePolicy struct {
	Limit   int
	Window  time.Duration
	MaxKeys int

	mu      sync.Mutex
	windows map[string]*rateWindow
}

type rateWindow struct {
	start time.Time
	count int
}

// Allow admits one request for the key inside the current window.
func (p *RatePolicy) Allow(key string, now time.Time) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.windows == nil {
		p.windows = make(map[string]*rateWindow)
	}
	w, ok := p.windows[key]
	if ok && now.Sub(w.start) >= p.Window {
		delete(p.windows, key)
		ok = false
	}
	if !ok {
		if len(p.windows) >= p.MaxKeys {
			p.evictExpired(now)
		}
		if len(p.windows) >= p.MaxKeys {
			return false // key budget exhausted: refuse, never grow
		}
		w = &rateWindow{start: now}
		p.windows[key] = w
	}
	if w.count >= p.Limit {
		return false
	}
	w.count++
	return true
}

func (p *RatePolicy) evictExpired(now time.Time) {
	for key, w := range p.windows {
		if now.Sub(w.start) >= p.Window {
			delete(p.windows, key)
		}
	}
}

// WithRate applies the policy per authenticated principal; anonymous
// requests share one bucket.
func (p *RatePolicy) WithRate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := "anonymous"
		if pr, ok := PrincipalFrom(r.Context()); ok {
			key = pr.InstallationID
		}
		if !p.Allow(key, time.Now()) {
			w.Header().Set("Retry-After",
				strconv.Itoa(int(p.Window/time.Second)+1))
			WriteError(w, r, http.StatusTooManyRequests, "rate_limited",
				"request rate exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}
