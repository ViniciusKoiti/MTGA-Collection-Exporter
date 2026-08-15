package consent

import (
	"fmt"
	"strconv"
)

// maxValueLen bounds every attribute value on the wire.
const maxValueLen = 128

// Admit accepts one event instance only when the event is declared,
// the installation granted the CURRENT purpose version, and every
// attribute is declared, typed and bounded. Anything else is refusal.
func (a Allowlist) Admit(event string, attrs map[string]string,
	granted map[string]int) error {
	spec, ok := a.events[event]
	if !ok {
		return fmt.Errorf("consent: event %q outside the allowlist", event)
	}
	current := a.purposes[spec.Purpose]
	if granted[spec.Purpose] != current {
		return fmt.Errorf("consent: purpose %q requires version %d consent",
			spec.Purpose, current)
	}
	for name, value := range attrs {
		kind, declared := spec.Attrs[name]
		if !declared {
			return fmt.Errorf("consent: attr %q not declared for %s", name, event)
		}
		if len(value) > maxValueLen {
			return fmt.Errorf("consent: attr %q oversized", name)
		}
		if err := checkType(kind, value); err != nil {
			return fmt.Errorf("consent: attr %q: %w", name, err)
		}
		if kind == AttrString {
			if err := Screen(value); err != nil {
				return fmt.Errorf("consent: attr %q: %w", name, err)
			}
		}
	}
	return nil
}

// Retention answers the retention class of a declared event.
func (a Allowlist) Retention(event string) (RetentionClass, error) {
	spec, ok := a.events[event]
	if !ok {
		return "", fmt.Errorf("consent: event %q outside the allowlist", event)
	}
	return spec.Retention, nil
}

func checkType(kind AttrType, value string) error {
	switch kind {
	case AttrInt:
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("expected int, got %q", value)
		}
	case AttrBool:
		if value != "true" && value != "false" {
			return fmt.Errorf("expected bool, got %q", value)
		}
	}
	return nil
}
