package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRatePolicyRefusesBeyondLimitAndResets(t *testing.T) {
	p := &RatePolicy{Limit: 2, Window: time.Minute, MaxKeys: 4}
	now := time.Unix(1000, 0)
	if !p.Allow("inst-1", now) || !p.Allow("inst-1", now) {
		t.Fatal("requests within the limit must be admitted")
	}
	if p.Allow("inst-1", now) {
		t.Fatal("request beyond the limit must be refused")
	}
	if !p.Allow("inst-1", now.Add(time.Minute)) {
		t.Fatal("a new window must admit again")
	}
}

func TestRatePolicyKeyBudgetIsBounded(t *testing.T) {
	p := &RatePolicy{Limit: 1, Window: time.Minute, MaxKeys: 2}
	now := time.Unix(1000, 0)
	if !p.Allow("a", now) || !p.Allow("b", now) {
		t.Fatal("keys within the budget must be admitted")
	}
	if p.Allow("c", now) {
		t.Fatal("a key beyond the budget must be refused, not grown")
	}
	if !p.Allow("c", now.Add(2*time.Minute)) {
		t.Fatal("expired keys must be evicted to admit new ones")
	}
}

func TestWithRateEmitsStableRefusal(t *testing.T) {
	p := &RatePolicy{Limit: 1, Window: time.Minute, MaxKeys: 2}
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), WithRequestID, p.WithRate)
	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	second := httptest.NewRecorder()
	h.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/", nil))
	if first.Code != http.StatusOK || second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 200 then 429, got %d then %d",
			first.Code, second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("rate refusal must carry Retry-After")
	}
}
