package testkit

import (
	"errors"
	"net/http"
	"testing"
)

// dialCanary fails the test if a refused request ever reaches it.
type dialCanary struct{ t *testing.T }

func (c dialCanary) RoundTrip(req *http.Request) (*http.Response, error) {
	c.t.Fatalf("refused request left the process: %s", req.URL)
	return nil, errors.New("unreachable")
}

func TestTrustProfileSetIsClosed(t *testing.T) {
	want := []string{"deterministic", "integration", "end-to-end",
		"staging", "provider-smoke"}
	profiles := Profiles()
	if len(profiles) != len(want) {
		t.Fatalf("the profile set must be exactly the five tiers: %v", profiles)
	}
	for _, name := range want {
		if _, err := ProfileByName(name); err != nil {
			t.Fatalf("tier %s must exist: %v", name, err)
		}
	}
	if _, err := ProfileByName("yolo"); err == nil {
		t.Fatal("unknown tiers must be refused")
	}
}

func TestDeterministicAndIntegrationAllowNoNetwork(t *testing.T) {
	for _, name := range []string{"deterministic", "integration"} {
		profile, _ := ProfileByName(name)
		if profile.AllowsRealClock {
			t.Fatalf("%s must use the controlled clock", name)
		}
		for _, host := range []string{"api.scryfall.com", "localhost",
			"127.0.0.1", "central.staging.invalid"} {
			if err := profile.AllowHost(host); err == nil {
				t.Fatalf("%s must not reach %s", name, host)
			}
		}
	}
}

func TestProviderSmokeReachesOnlyTheProvider(t *testing.T) {
	profile, _ := ProfileByName("provider-smoke")
	if err := profile.AllowHost("api.scryfall.com"); err != nil {
		t.Fatalf("the provider host must be allowed: %v", err)
	}
	if err := profile.AllowHost("central.staging.invalid"); err == nil {
		t.Fatal("provider smoke must not reach staging")
	}
}

// TestGuardedTransportRefusesBeforeDialing: the canary proves refused
// requests never leave the process.
func TestGuardedTransportRefusesBeforeDialing(t *testing.T) {
	profile, _ := ProfileByName("deterministic")
	transport := GuardedTransport{Profile: profile, Next: dialCanary{t}}
	req, err := http.NewRequest(http.MethodGet, "https://api.scryfall.com/x", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if _, err := transport.RoundTrip(req); err == nil {
		t.Fatal("the deterministic tier must refuse every host")
	}
}
