package catalogclient

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"
)

type memCache struct {
	man    Manifest
	object []byte
	ok     bool
}

func (c *memCache) Store(man Manifest, object []byte) error {
	c.man, c.object, c.ok = man, append([]byte(nil), object...), true
	return nil
}

func (c *memCache) Load() (Manifest, []byte, bool) {
	return c.man, c.object, c.ok
}

type memFetcher struct {
	man     Manifest
	object  []byte
	offline bool
}

func (f memFetcher) Manifest(context.Context) (Manifest, error) {
	if f.offline {
		return Manifest{}, errors.New("network unreachable")
	}
	return f.man, nil
}

func (f memFetcher) Object(context.Context, string) ([]byte, error) {
	if f.offline {
		return nil, errors.New("network unreachable")
	}
	return f.object, nil
}

// signedFixture builds a wire-faithful signed manifest for the object.
func signedFixture(t *testing.T, object []byte) (Manifest, Verifier) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	sum := sha256.Sum256(object)
	ref := ObjectRef{Key: "catalog/fixture.json.gz",
		SHA256: hex.EncodeToString(sum[:]), Size: int64(len(object))}
	man := Manifest{KeyID: "k1", Schema: "cards-v1", Object: ref,
		Signature: base64.StdEncoding.EncodeToString(
			ed25519.Sign(private, canonicalPayload(ref)))}
	return man, Verifier{Trusted: map[string]ed25519.PublicKey{"k1": public}}
}

// TestDesktopVerifiesThenFallsBackOffline: a verified catalog is
// cached, and when the network dies the cached one serves, marked
// stale — the offline fallback of central task 4.7.
func TestDesktopVerifiesThenFallsBackOffline(t *testing.T) {
	object := []byte("fixture-catalog-bytes")
	man, verifier := signedFixture(t, object)
	cache := &memCache{}
	online := Client{Fetch: memFetcher{man: man, object: object},
		Cache: cache, Verify: verifier, Schema: "cards-v1"}
	fresh, err := online.Current(context.Background())
	if err != nil || fresh.Stale {
		t.Fatalf("verified catalog must serve fresh: %+v %v", fresh, err)
	}
	offline := Client{Fetch: memFetcher{offline: true},
		Cache: cache, Verify: verifier, Schema: "cards-v1"}
	cached, err := offline.Current(context.Background())
	if err != nil || !cached.Stale || string(cached.Object) != string(object) {
		t.Fatalf("offline must fall back to the cached catalog: %+v %v",
			cached, err)
	}
	empty := Client{Fetch: memFetcher{offline: true},
		Cache: &memCache{}, Verify: verifier, Schema: "cards-v1"}
	if _, err := empty.Current(context.Background()); err == nil {
		t.Fatal("offline with no cache must be an error, not a guess")
	}
}
