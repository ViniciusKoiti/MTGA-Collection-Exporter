package collectionsvc

import (
	"sort"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// Change is one card whose count differs between two snapshots;
// Before == 0 means added, After == 0 means removed.
type Change struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	Set    string `json:"set"`
	Before int    `json:"before"`
	After  int    `json:"after"`
}

// Diff is the snapshot comparison the Collection view renders.
type Diff struct {
	TotalBefore int      `json:"total_before"`
	TotalAfter  int      `json:"total_after"`
	Changes     []Change `json:"changes"`
}

func keyOf(entry collection.Entry) string {
	if entry.Identity.Printing != "" {
		return string(entry.Identity.Printing)
	}
	return "raw:" + entry.Raw
}

// Compare diffs two snapshots by printing identity (raw text for
// unresolved records), sorted by name then key — deterministic.
func Compare(previous, latest collection.Snapshot) Diff {
	type side struct {
		entry collection.Entry
		count int
	}
	byKey := map[string]*[2]side{}
	index := func(snap collection.Snapshot, slot int, total *int) {
		for _, entry := range snap.Entries {
			*total += entry.Quantity
			key := keyOf(entry)
			if byKey[key] == nil {
				byKey[key] = &[2]side{}
			}
			byKey[key][slot] = side{entry: entry,
				count: byKey[key][slot].count + entry.Quantity}
		}
	}
	diff := Diff{}
	index(previous, 0, &diff.TotalBefore)
	index(latest, 1, &diff.TotalAfter)
	for key, sides := range byKey {
		if sides[0].count == sides[1].count {
			continue
		}
		entry := sides[1].entry
		if entry.Quantity == 0 && sides[0].count > 0 {
			entry = sides[0].entry
		}
		diff.Changes = append(diff.Changes, Change{Key: key,
			Name: entry.Identity.Name, Set: entry.Identity.Set,
			Before: sides[0].count, After: sides[1].count})
	}
	sort.Slice(diff.Changes, func(i, j int) bool {
		if diff.Changes[i].Name != diff.Changes[j].Name {
			return diff.Changes[i].Name < diff.Changes[j].Name
		}
		return diff.Changes[i].Key < diff.Changes[j].Key
	})
	return diff
}
