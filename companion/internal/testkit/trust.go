package testkit

import (
	"fmt"
	"net/http"
)

// TrustProfile is one execution tier of the harness (OpenSpec
// add-graph-workflow-harness, task 5.7): it fixes what the tier may
// reach on the network and whether wall-clock time is allowed. The
// set of profiles is closed — a suite cannot invent its own tier.
type TrustProfile struct {
	Name            string
	AllowedHosts    []string // empty means NO network at all
	AllowsRealClock bool
}

// Profiles returns the closed set of trust profiles.
func Profiles() map[string]TrustProfile {
	return map[string]TrustProfile{
		"deterministic": {Name: "deterministic"},
		"integration":   {Name: "integration"},
		"end-to-end": {Name: "end-to-end", AllowsRealClock: true,
			AllowedHosts: []string{"127.0.0.1", "localhost"}},
		"staging": {Name: "staging", AllowsRealClock: true,
			AllowedHosts: []string{"central.staging.invalid"}},
		"provider-smoke": {Name: "provider-smoke", AllowsRealClock: true,
			AllowedHosts: []string{"api.scryfall.com"}},
	}
}

// ProfileByName resolves a tier or refuses unknown names.
func ProfileByName(name string) (TrustProfile, error) {
	profile, ok := Profiles()[name]
	if !ok {
		return TrustProfile{}, fmt.Errorf("testkit: unknown trust profile %q", name)
	}
	return profile, nil
}

// AllowHost answers whether the profile may reach the host.
func (p TrustProfile) AllowHost(host string) error {
	for _, allowed := range p.AllowedHosts {
		if host == allowed {
			return nil
		}
	}
	return fmt.Errorf("testkit: profile %q must not reach host %q", p.Name, host)
}

// GuardedTransport enforces the allowlist BEFORE any dial: a refused
// request never leaves the process.
type GuardedTransport struct {
	Profile TrustProfile
	Next    http.RoundTripper
}

// RoundTrip refuses disallowed hosts and delegates the rest.
func (g GuardedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := g.Profile.AllowHost(req.URL.Hostname()); err != nil {
		return nil, err
	}
	next := g.Next
	if next == nil {
		next = http.DefaultTransport
	}
	return next.RoundTrip(req)
}
