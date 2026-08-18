package decks

import "strings"

// legalIn reports whether the card is legal in the requested format.
func legalIn(card CardProfile, format string) bool {
	for _, f := range card.Formats {
		if strings.EqualFold(f, format) {
			return true
		}
	}
	return false
}

// overlaps reports whether the two type lists share any entry.
func overlaps(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if strings.EqualFold(x, y) {
				return true
			}
		}
	}
	return false
}

// sameSet reports whether the two color lists are equal as sets,
// case-insensitively; color identity must match exactly for the
// color_match evidence.
func sameSet(a, b []string) bool {
	if len(normalize(a)) != len(normalize(b)) {
		return false
	}
	setA := normalize(a)
	for color := range normalize(b) {
		if !setA[color] {
			return false
		}
	}
	return true
}

func normalize(colors []string) map[string]bool {
	set := make(map[string]bool, len(colors))
	for _, c := range colors {
		set[strings.ToLower(c)] = true
	}
	return set
}
