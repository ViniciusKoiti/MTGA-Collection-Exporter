package consent

import (
	"strings"
	"testing"
)

func productAllowlist(t *testing.T) Allowlist {
	t.Helper()
	a, err := New(
		[]Purpose{{Name: "product", Version: 2}},
		[]EventSpec{{Name: "scan_completed", Purpose: "product",
			Retention: RetentionShort,
			Attrs: map[string]AttrType{"result": AttrString,
				"cards": AttrInt, "offline": AttrBool}}})
	if err != nil {
		t.Fatalf("allowlist: %v", err)
	}
	return a
}

func TestConstructionRefusesUnboundedDeclarations(t *testing.T) {
	purposes := []Purpose{{Name: "product", Version: 1}}
	for name, event := range map[string]EventSpec{
		"unknown purpose": {Name: "e", Purpose: "ghost",
			Retention: RetentionShort},
		"no retention": {Name: "e", Purpose: "product"},
		"too many attrs": {Name: "e", Purpose: "product",
			Retention: RetentionShort, Attrs: map[string]AttrType{
				"a1": AttrString, "a2": AttrString, "a3": AttrString,
				"a4": AttrString, "a5": AttrString, "a6": AttrString,
				"a7": AttrString, "a8": AttrString, "a9": AttrString}},
		"bad attr type": {Name: "e", Purpose: "product",
			Retention: RetentionShort,
			Attrs:     map[string]AttrType{"a": AttrType("blob")}},
	} {
		if _, err := New(purposes, []EventSpec{event}); err == nil {
			t.Fatalf("%s must be refused at construction", name)
		}
	}
	if _, err := New([]Purpose{{Name: "p", Version: 0}}, nil); err == nil {
		t.Fatal("versionless purposes must be refused")
	}
}

func TestAdmitDemandsCurrentConsentAndDeclaredAttrs(t *testing.T) {
	a := productAllowlist(t)
	good := map[string]string{"result": "ok", "cards": "1200", "offline": "false"}
	if err := a.Admit("scan_completed", good,
		map[string]int{"product": 2}); err != nil {
		t.Fatalf("declared event with current consent must pass: %v", err)
	}
	cases := map[string]struct {
		event   string
		attrs   map[string]string
		granted map[string]int
	}{
		"outside allowlist": {"raw_log", nil, map[string]int{"product": 2}},
		"no consent":        {"scan_completed", good, nil},
		"stale consent":     {"scan_completed", good, map[string]int{"product": 1}},
		"undeclared attr": {"scan_completed",
			map[string]string{"deck": "x"}, map[string]int{"product": 2}},
		"type mismatch": {"scan_completed",
			map[string]string{"cards": "many"}, map[string]int{"product": 2}},
		"oversized value": {"scan_completed",
			map[string]string{"result": strings.Repeat("x", 200)},
			map[string]int{"product": 2}},
	}
	for name, c := range cases {
		if err := a.Admit(c.event, c.attrs, c.granted); err == nil {
			t.Fatalf("%s must be refused", name)
		}
	}
}

func TestRetentionClassIsFixedAtDeclaration(t *testing.T) {
	a := productAllowlist(t)
	class, err := a.Retention("scan_completed")
	if err != nil || class != RetentionShort {
		t.Fatalf("declared retention must be answered: %v %v", class, err)
	}
	if _, err := a.Retention("raw_log"); err == nil {
		t.Fatal("undeclared events have no retention to answer")
	}
}
