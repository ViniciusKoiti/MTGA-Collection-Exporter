package catalogclient

import (
	"context"
	"testing"
)

// TestVerificationFailuresNeverPoisonTheCache: tampered objects,
// unknown keys and incompatible schemas all fall back to the cached
// verified catalog — nothing unverified is ever stored or served.
func TestVerificationFailuresNeverPoisonTheCache(t *testing.T) {
	object := []byte("fixture-catalog-bytes")
	man, verifier := signedFixture(t, object)
	cache := &memCache{}
	good := Client{Fetch: memFetcher{man: man, object: object},
		Cache: cache, Verify: verifier, Schema: "cards-v1"}
	if _, err := good.Current(context.Background()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	cases := map[string]Client{
		"tampered object": {Fetch: memFetcher{man: man,
			object: []byte("evil-bytes")}, Cache: cache,
			Verify: verifier, Schema: "cards-v1"},
		"unknown key": {Fetch: memFetcher{man: man, object: object},
			Cache: cache, Verify: Verifier{}, Schema: "cards-v1"},
		"incompatible schema": {Fetch: memFetcher{man: man, object: object},
			Cache: cache, Verify: verifier, Schema: "cards-v2"},
	}
	for name, client := range cases {
		result, err := client.Current(context.Background())
		if err != nil || !result.Stale {
			t.Fatalf("%s: must fall back to the cached catalog: %+v %v",
				name, result, err)
		}
		if string(result.Object) != string(object) {
			t.Fatalf("%s: the cache must stay the verified bytes", name)
		}
	}
}
