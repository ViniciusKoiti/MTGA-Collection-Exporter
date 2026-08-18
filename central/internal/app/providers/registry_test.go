package providers

import (
	"errors"
	"testing"
	"time"
)

func scryfallOnly(t *testing.T, perMinute int) *Registry {
	t.Helper()
	reg, err := NewRegistry(Provider{Name: "scryfall",
		Rights: Rights{AttributionRequired: true,
			AllowedKinds: []string{"cards"}, TermsURL: "https://scryfall.com"},
		RequestsPerMinute: perMinute})
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	return reg
}

func TestRegistryRefusesIncompleteApprovals(t *testing.T) {
	for name, p := range map[string]Provider{
		"unnamed":    {RequestsPerMinute: 1, Rights: Rights{AllowedKinds: []string{"cards"}}},
		"unbudgeted": {Name: "x", Rights: Rights{AllowedKinds: []string{"cards"}}},
		"rightless":  {Name: "x", RequestsPerMinute: 1},
	} {
		if _, err := NewRegistry(p); err == nil {
			t.Fatalf("%s approval must be refused", name)
		}
	}
	dup := Provider{Name: "x", RequestsPerMinute: 1,
		Rights: Rights{AllowedKinds: []string{"cards"}}}
	if _, err := NewRegistry(dup, dup); err == nil {
		t.Fatal("duplicate approvals must be refused")
	}
}

func TestAuthorizeEnforcesRightsAndRate(t *testing.T) {
	reg := scryfallOnly(t, 2)
	now := time.Unix(1000, 0)
	if _, err := reg.Authorize("gatherer", "cards", now); !errors.Is(err, ErrNotApproved) {
		t.Fatalf("unapproved provider: %v", err)
	}
	if _, err := reg.Authorize("scryfall", "meta", now); !errors.Is(err, ErrKindNotAllowed) {
		t.Fatalf("kind outside rights: %v", err)
	}
	for i := range 2 {
		if _, err := reg.Authorize("scryfall", "cards", now); err != nil {
			t.Fatalf("request %d within budget: %v", i, err)
		}
	}
	if _, err := reg.Authorize("scryfall", "cards", now); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("budget spent must refuse: %v", err)
	}
	if _, err := reg.Authorize("scryfall", "cards", now.Add(time.Minute)); err != nil {
		t.Fatalf("new window must admit: %v", err)
	}
}

func TestDisablementIsImmediateAndReversible(t *testing.T) {
	reg := scryfallOnly(t, 10)
	now := time.Unix(1000, 0)
	reg.Disable("scryfall")
	if _, err := reg.Authorize("scryfall", "cards", now); !errors.Is(err, ErrDisabled) {
		t.Fatalf("disabled provider must refuse: %v", err)
	}
	reg.Enable("scryfall")
	if _, err := reg.Authorize("scryfall", "cards", now); err != nil {
		t.Fatalf("re-enabled provider must admit: %v", err)
	}
}
