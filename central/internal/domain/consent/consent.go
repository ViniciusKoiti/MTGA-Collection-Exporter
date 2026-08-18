// Package consent defines purpose-versioned consent and the closed
// telemetry event allowlist (OpenSpec add-central-go-platform, task
// 5.1). Everything here is bounded at declaration time: names,
// attribute sets, value types and retention classes.
package consent

import "fmt"

// Purpose is one consent purpose at its current version; bumping the
// version voids every earlier grant for that purpose.
type Purpose struct {
	Name    string
	Version int
}

// RetentionClass bounds how long an accepted event may live.
type RetentionClass string

const (
	RetentionEphemeral RetentionClass = "ephemeral" // hours, never backed up
	RetentionShort     RetentionClass = "short"     // days, purged by TTL
	RetentionAggregate RetentionClass = "aggregate" // survives only in rollups
)

// AttrType is the closed set of value types an attribute may carry.
type AttrType string

const (
	AttrString AttrType = "string"
	AttrInt    AttrType = "int"
	AttrBool   AttrType = "bool"
)

// EventSpec declares one allowlisted event: purpose, retention class
// and a closed attribute set, all fixed at declaration.
type EventSpec struct {
	Name      string
	Purpose   string
	Retention RetentionClass
	Attrs     map[string]AttrType
}

// Allowlist is the closed event vocabulary bound to purposes.
type Allowlist struct {
	purposes map[string]int
	events   map[string]EventSpec
}

// New validates every bound once, at construction.
func New(purposes []Purpose, events []EventSpec) (Allowlist, error) {
	a := Allowlist{purposes: map[string]int{}, events: map[string]EventSpec{}}
	for _, p := range purposes {
		if p.Name == "" || p.Version < 1 {
			return Allowlist{}, fmt.Errorf("consent: invalid purpose %+v", p)
		}
		if _, dup := a.purposes[p.Name]; dup {
			return Allowlist{}, fmt.Errorf("consent: duplicate purpose %s", p.Name)
		}
		a.purposes[p.Name] = p.Version
	}
	for _, e := range events {
		if err := a.checkSpec(e); err != nil {
			return Allowlist{}, err
		}
		a.events[e.Name] = e
	}
	return a, nil
}

func (a Allowlist) checkSpec(e EventSpec) error {
	if e.Name == "" || len(e.Name) > 64 || len(e.Attrs) > 8 {
		return fmt.Errorf("consent: unbounded event spec %q", e.Name)
	}
	if _, dup := a.events[e.Name]; dup {
		return fmt.Errorf("consent: duplicate event %s", e.Name)
	}
	if _, ok := a.purposes[e.Purpose]; !ok {
		return fmt.Errorf("consent: event %s names unknown purpose %q",
			e.Name, e.Purpose)
	}
	switch e.Retention {
	case RetentionEphemeral, RetentionShort, RetentionAggregate:
	default:
		return fmt.Errorf("consent: event %s has no retention class", e.Name)
	}
	for name, kind := range e.Attrs {
		if name == "" || len(name) > 32 ||
			(kind != AttrString && kind != AttrInt && kind != AttrBool) {
			return fmt.Errorf("consent: event %s attr %q unbounded", e.Name, name)
		}
	}
	return nil
}
