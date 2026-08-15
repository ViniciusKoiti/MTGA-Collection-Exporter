package httpapi

import (
	"net/http"
	"time"
)

// Observation is one low-cardinality request record: route pattern and
// status only, never raw paths, tokens or identifiers.
type Observation struct {
	Route    string
	Status   int
	Duration time.Duration
}

// statusRecorder captures the status written by the handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// WithObserver reports one observation per request to the sink; the
// route label must be the registered pattern, not the raw URL.
func WithObserver(route string,
	sink func(Observation)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rec, r)
			sink(Observation{Route: route, Status: rec.status,
				Duration: time.Since(start)})
		})
	}
}
