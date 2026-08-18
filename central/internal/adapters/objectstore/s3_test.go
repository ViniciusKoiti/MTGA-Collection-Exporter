package objectstore

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

// fakeS3 is an in-process S3-compatible object endpoint.
type fakeS3 struct {
	mu      sync.Mutex
	objects map[string][]byte
	tamper  bool // corrupt reads to trip verification
	deletes []string
}

func (f *fakeS3) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		key := strings.TrimPrefix(r.URL.Path, "/")
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			f.objects[key] = body
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			body, ok := f.objects[key]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if f.tamper {
				body = append([]byte("x"), body...)
			}
			_, _ = w.Write(body)
		case http.MethodDelete:
			delete(f.objects, key)
			f.deletes = append(f.deletes, key)
			w.WriteHeader(http.StatusNoContent)
		}
	})
}

func rig(t *testing.T) (*fakeS3, S3Writer) {
	t.Helper()
	fake := &fakeS3{objects: map[string][]byte{}}
	server := httptest.NewServer(fake.handler())
	t.Cleanup(server.Close)
	return fake, S3Writer{Endpoint: server.URL, Bucket: "catalogs"}
}

func snapshot() catalog.Snapshot {
	return catalog.Snapshot{SchemaVersion: "v1", Cards: []catalog.Card{
		{GrpID: 1, Name: "Synthetic Bolt", Set: "TST"},
		{GrpID: 2, Name: "Synthetic Counter", Set: "TST"}}}
}

// TestWriteSnapshotIsContentAddressedAndVerified: same content, same
// immutable key; the reference carries the verified hash and size.
func TestWriteSnapshotIsContentAddressedAndVerified(t *testing.T) {
	fake, writer := rig(t)
	first, err := writer.WriteSnapshot(context.Background(), snapshot())
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.HasPrefix(first.Key, "catalog/") ||
		!strings.Contains(first.Key, first.SHA256) {
		t.Fatalf("keys must be content-addressed: %+v", first)
	}
	again, err := writer.WriteSnapshot(context.Background(), snapshot())
	if err != nil || again.Key != first.Key {
		t.Fatalf("same content must map to the same immutable key: %+v %v",
			again, err)
	}
	if len(fake.objects) != 1 || first.Size == 0 {
		t.Fatalf("exactly one object must exist: %d", len(fake.objects))
	}
}

// TestFailedVerificationCleansUpAndErrors: tampered storage fails the
// upload and the broken object is deleted, never left published.
func TestFailedVerificationCleansUpAndErrors(t *testing.T) {
	fake, writer := rig(t)
	fake.tamper = true
	if _, err := writer.WriteSnapshot(context.Background(), snapshot()); err == nil {
		t.Fatal("a verification mismatch must fail the upload")
	}
	if len(fake.deletes) != 1 || len(fake.objects) != 0 {
		t.Fatalf("the broken object must be cleaned up: deletes=%v objects=%d",
			fake.deletes, len(fake.objects))
	}
}
