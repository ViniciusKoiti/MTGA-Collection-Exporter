// Package collectionsvc powers the Collection view (OpenSpec
// introduce-agentic-go-companion, task 4.4): search, sorting, compact
// compound filters, per-card details, unresolved records and snapshot
// comparison — pure functions over the domain snapshot.
package collectionsvc

import (
	"sort"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// Query is the compound filter of the Collection view; every field
// composes with the others.
type Query struct {
	Text           string `json:"text"`
	Set            string `json:"set"`
	UnresolvedOnly bool   `json:"unresolved_only"`
	MinQuantity    int    `json:"min_quantity"`
	Sort           string `json:"sort"` // name | set | quantity
}

// Row is one rendered collection line, details included.
type Row struct {
	Printing   string `json:"printing"`
	Name       string `json:"name"`
	Set        string `json:"set"`
	Quantity   int    `json:"quantity"`
	Unresolved bool   `json:"unresolved"`
	Raw        string `json:"raw,omitempty"`
	Oracle     string `json:"oracle,omitempty"`
	Arena      int    `json:"arena,omitempty"`
}

func rowOf(entry collection.Entry) Row {
	return Row{
		Printing:   string(entry.Identity.Printing),
		Name:       entry.Identity.Name,
		Set:        entry.Identity.Set,
		Quantity:   entry.Quantity,
		Unresolved: entry.Unresolved,
		Raw:        entry.Raw,
		Oracle:     string(entry.Identity.Oracle),
		Arena:      int(entry.Identity.Arena),
	}
}

func matches(row Row, q Query) bool {
	if q.UnresolvedOnly && !row.Unresolved {
		return false
	}
	if q.Set != "" && !strings.EqualFold(row.Set, q.Set) {
		return false
	}
	if row.Quantity < q.MinQuantity {
		return false
	}
	if q.Text == "" {
		return true
	}
	needle := strings.ToLower(q.Text)
	return strings.Contains(strings.ToLower(row.Name), needle) ||
		strings.Contains(strings.ToLower(row.Raw), needle)
}

// Page filters and sorts the snapshot deterministically; an unknown
// sort key falls back to name, and ties always break the same way.
func Page(snap collection.Snapshot, q Query) []Row {
	rows := make([]Row, 0, len(snap.Entries))
	for _, entry := range snap.Entries {
		if row := rowOf(entry); matches(row, q) {
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		switch q.Sort {
		case "set":
			if a.Set != b.Set {
				return a.Set < b.Set
			}
		case "quantity":
			if a.Quantity != b.Quantity {
				return a.Quantity > b.Quantity
			}
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		if a.Set != b.Set {
			return a.Set < b.Set
		}
		return a.Printing < b.Printing
	})
	return rows
}
