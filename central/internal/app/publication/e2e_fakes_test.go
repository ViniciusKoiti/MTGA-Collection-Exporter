package publication_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

// quarantineFail fails the test if any fixture card is quarantined.
type quarantineFail struct{ t *testing.T }

func (q quarantineFail) Quarantine(_ context.Context, card catalog.Card,
	reason string) error {
	q.t.Fatalf("fixture card quarantined: %+v (%s)", card, reason)
	return nil
}

// e2eS3 is the in-process S3-compatible endpoint of the e2e test.
type e2eS3 struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func newE2ES3() (*e2eS3, *httptest.Server) {
	fake := &e2eS3{objects: map[string][]byte{}}
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fake.mu.Lock()
			defer fake.mu.Unlock()
			key := r.URL.Path[1:]
			switch r.Method {
			case http.MethodPut:
				body, _ := io.ReadAll(r.Body)
				fake.objects[key] = body
			case http.MethodGet:
				body, ok := fake.objects[key]
				if !ok {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				_, _ = w.Write(body)
			case http.MethodDelete:
				delete(fake.objects, key)
				w.WriteHeader(http.StatusNoContent)
			}
		}))
	return fake, server
}

func (f *e2eS3) object(key string) ([]byte, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	body, ok := f.objects["catalogs/"+key]
	return body, ok
}

// e2eDeps provides fixture cards and records batches and activations.
type e2eDeps struct {
	mu          sync.Mutex
	activations []catalog.Manifest
	batches     int
}

func (d *e2eDeps) Fetch(_ context.Context,
	job catalog.ProviderJob) ([]catalog.Observation, error) {
	obs := make([]catalog.Observation, 5)
	for i := range obs {
		obs[i] = catalog.Observation{Provider: job.Provider, GrpID: 1001 + i}
	}
	return obs, nil
}

func (d *e2eDeps) Normalize(_ context.Context,
	obs catalog.Observation) (catalog.Card, error) {
	return catalog.Card{GrpID: obs.GrpID,
		Name: fmt.Sprintf("Fixture Card %d", obs.GrpID), Set: "FIX"}, nil
}

func (d *e2eDeps) WriteBatch(_ context.Context, _ []catalog.Card) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.batches++
	return nil
}

func (d *e2eDeps) Activate(_ context.Context, man catalog.Manifest) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.activations = append(d.activations, man)
	return nil
}
