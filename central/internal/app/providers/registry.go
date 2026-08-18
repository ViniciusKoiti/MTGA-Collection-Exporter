// Package providers governs which external catalog providers may be
// used at all: approval, usage rights, per-provider rate budgets and
// rapid disablement (OpenSpec add-central-go-platform, task 4.1).
package providers

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

// Rights records what the provider's terms actually allow.
type Rights struct {
	AttributionRequired bool
	AllowedKinds        []string
	TermsURL            string
}

// Provider is one approved upstream with its rights and rate budget.
type Provider struct {
	Name              string
	Rights            Rights
	RequestsPerMinute int
}

// Registry is the closed list of approved providers; anything not
// registered simply cannot be fetched.
type Registry struct {
	mu        sync.Mutex
	providers map[string]Provider
	disabled  map[string]bool
	windows   map[string]*rateWindow
}

type rateWindow struct {
	start time.Time
	count int
}

// NewRegistry refuses duplicate, unnamed or unbudgeted providers.
func NewRegistry(approved ...Provider) (*Registry, error) {
	reg := &Registry{providers: map[string]Provider{},
		disabled: map[string]bool{}, windows: map[string]*rateWindow{}}
	for _, p := range approved {
		if p.Name == "" || p.RequestsPerMinute < 1 ||
			len(p.Rights.AllowedKinds) == 0 {
			return nil, fmt.Errorf("providers: incomplete approval: %+v", p)
		}
		if _, dup := reg.providers[p.Name]; dup {
			return nil, fmt.Errorf("providers: duplicate approval: %s", p.Name)
		}
		reg.providers[p.Name] = p
	}
	return reg, nil
}

// Authorize admits one request for the provider and kind at the given
// instant, consuming one slot of its rate budget.
func (r *Registry) Authorize(name, kind string, now time.Time) (Provider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.providers[name]
	if !ok {
		return Provider{}, ErrNotApproved
	}
	if r.disabled[name] {
		return Provider{}, ErrDisabled
	}
	if !slices.Contains(p.Rights.AllowedKinds, kind) {
		return Provider{}, ErrKindNotAllowed
	}
	w := r.windows[name]
	if w == nil || now.Sub(w.start) >= time.Minute {
		w = &rateWindow{start: now}
		r.windows[name] = w
	}
	if w.count >= p.RequestsPerMinute {
		return Provider{}, ErrRateLimited
	}
	w.count++
	return p, nil
}

// Disable cuts the provider off immediately; Enable restores it.
func (r *Registry) Disable(name string) { r.setDisabled(name, true) }

// Enable restores a previously disabled provider.
func (r *Registry) Enable(name string) { r.setDisabled(name, false) }

func (r *Registry) setDisabled(name string, value bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.disabled[name] = value
}
