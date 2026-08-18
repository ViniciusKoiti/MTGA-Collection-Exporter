package arch

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func frontendSource(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "desktop",
		"frontend", "src", name))
	if err != nil {
		t.Fatalf("frontend source %s: %v", name, err)
	}
	return string(raw)
}

// TestPrimaryFlowsCarryAccessibilityContracts is the automated
// accessibility gate of task 4.6: the primary flows must keep their
// keyboard navigation, labels and text status — removing any of them
// fails here before a reviewer ever sees it.
func TestPrimaryFlowsCarryAccessibilityContracts(t *testing.T) {
	shell := frontendSource(t, "main.ts")
	for _, marker := range []string{`aria-label="Main navigation"`,
		"aria-current", "ArrowDown", "ArrowUp", "keydown"} {
		if !strings.Contains(shell, marker) {
			t.Errorf("main.ts lost its %q accessibility contract", marker)
		}
	}
	state := frontendSource(t, "state.ts")
	if !strings.Contains(state, `role="status"`) {
		t.Error("state banners must announce as text status (role=status)")
	}
	collection := frontendSource(t, "collection.ts")
	unlabeled := regexp.MustCompile(`<input(?:[^>]*)>`)
	for _, input := range unlabeled.FindAllString(collection, -1) {
		labeled := strings.Contains(input, "aria-label=") ||
			strings.Contains(input, `type="checkbox"`) // wrapped by <label>
		if !labeled {
			t.Errorf("collection.ts has an unlabeled input: %s", input)
		}
	}
}

// TestLayoutScalesWithUserFontSettings: no pixel font sizes and a
// visible focus style — the layout must follow the user's settings.
func TestLayoutScalesWithUserFontSettings(t *testing.T) {
	css := frontendSource(t, "style.css")
	if pixelFonts := regexp.MustCompile(`font-size:\s*\d+px`).
		FindAllString(css, -1); len(pixelFonts) > 0 {
		t.Errorf("pixel font sizes do not scale: %v", pixelFonts)
	}
	if !strings.Contains(css, ":focus-visible") {
		t.Error("visible focus styling is non-negotiable")
	}
}
