package consent

import "testing"

func TestScreenRefusesDisallowedContentShapes(t *testing.T) {
	refused := map[string]string{
		"raw log":         "[UnityCrossThreadLogger]\nDeck list follows",
		"deck list":       "4 Lightning Bolt\n4 Counterspell",
		"windows path":    `C:\Users\someone\AppData\output_log.txt`,
		"unix path":       "/home/someone/.config/mtga/collection.json",
		"log extension":   "output_log.log",
		"bearer token":    "Bearer abc123def456",
		"jwt":             "eyJhbGciOiJIUzI1NiJ9.payload.sig",
		"embedded secret": "password=hunter2",
		"hex identity":    "3f9c2a1b4d5e6f708192a3b4",
		"mtga account id": "584930128476512",
		"prompt prose":    "please build me a deck that beats the meta today",
	}
	for name, value := range refused {
		if err := Screen(value); err == nil {
			t.Fatalf("%s must be refused: %q", name, value)
		}
	}
	allowed := []string{"ok", "json", "pt-BR", "cards-v1", "true",
		"1200", "scryfall", "export finished"}
	for _, value := range allowed {
		if err := Screen(value); err != nil {
			t.Fatalf("benign value %q must pass: %v", value, err)
		}
	}
}

// TestAdmitScreensDeclaredStringAttrs: the content gate runs inside
// Admit, so a declared attribute cannot smuggle a path or identity.
func TestAdmitScreensDeclaredStringAttrs(t *testing.T) {
	a := productAllowlist(t)
	granted := map[string]int{"product": 2}
	for name, value := range map[string]string{
		"path":     `C:\Users\x\collection.json`,
		"identity": "584930128476512",
		"log":      "line1\nline2",
	} {
		if err := a.Admit("scan_completed",
			map[string]string{"result": value}, granted); err == nil {
			t.Fatalf("declared attr carrying a %s must be refused", name)
		}
	}
}
