package publication_test

import (
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/adapters/objectstore"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/adapters/signing"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/app/publication"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/config"
)

// TestSignedFixtureCatalogRoundTrip publishes one fixture catalog
// through the REAL object-store and signing adapters, then plays the
// client: download by manifest key, hash check, Ed25519 verification
// and schema compatibility (task 4.7, central half; the desktop-side
// consumption stays open).
func TestSignedFixtureCatalogRoundTrip(t *testing.T) {
	fakeS3, server := newE2ES3()
	defer server.Close()
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	signer, err := signing.NewSigner("fixture-key", private)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}
	deps := &e2eDeps{}
	budgets := config.Budgets{FetchWorkers: 2, NormalizeWorkers: 2,
		ProviderBuffer: 2, ObservationBuffer: 4, NormalizedBuffer: 4,
		BatchSize: 3}
	man, err := publication.Run(t.Context(), budgets, "cards-v1",
		[]catalog.ProviderJob{{Provider: "fixture", Kind: "cards"}},
		publication.Deps{Fetcher: deps, Normalize: deps,
			Validator: publication.SoundCard{}, Quarantine: quarantineFail{t},
			Writer:  deps,
			Objects: objectstore.S3Writer{Endpoint: server.URL, Bucket: "catalogs"},
			Signer:  signer, Activator: deps})
	if err != nil {
		t.Fatalf("publication: %v", err)
	}
	if len(deps.activations) != 1 || deps.batches == 0 {
		t.Fatalf("exactly one activation expected: %+v", deps.activations)
	}
	verifier, err := signing.NewVerifier(
		map[string]ed25519.PublicKey{"fixture-key": public})
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	if err := verifier.Verify(man); err != nil {
		t.Fatalf("the client must accept the signature: %v", err)
	}
	body, ok := fakeS3.object(man.Object.Key)
	if !ok {
		t.Fatalf("download by manifest key failed: %s", man.Object.Key)
	}
	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != man.Object.SHA256 {
		t.Fatal("downloaded bytes must match the manifest hash")
	}
	zr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	raw, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var snap catalog.Snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("snapshot must decode: %v", err)
	}
	if snap.SchemaVersion != "cards-v1" || len(snap.Cards) != 5 {
		t.Fatalf("compatibility broke: %s %d", snap.SchemaVersion, len(snap.Cards))
	}
	tampered := man
	tampered.Object.SHA256 = "0000"
	if err := verifier.Verify(tampered); err == nil {
		t.Fatal("a tampered manifest must be rejected")
	}
}
