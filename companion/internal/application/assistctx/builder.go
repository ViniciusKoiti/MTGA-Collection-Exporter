// Package assistctx builds the provider-neutral context sent to the
// optional assistant model (OpenSpec introduce-agentic-go-companion,
// task 6.7). Only the data the conversation needs leaves the device:
// stable status codes, the deck under discussion and ownership
// aggregates — never raw logs, local paths, credentials, or collection
// entries unrelated to that deck. Output is deterministic and frozen by
// a redaction snapshot test.
package assistctx

import (
	"fmt"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

// maxQuestionLen caps the user question carried into the context.
const maxQuestionLen = 500

// Input is everything the builder may consider. Absent fields are simply
// omitted; there is no other data source.
type Input struct {
	SyncStatus string // stable code: fresh, stale, error, ...
	SnapshotID string
	TotalCards int
	Question   string
	Deck       *decks.Deck      // only the deck under discussion
	Ownership  *decks.Ownership // aggregate gaps for that deck
}

// Build renders the redacted context. Deck entries carry names and
// counts only; ownership contributes totals and gap names.
func Build(in Input) string {
	var b strings.Builder
	b.WriteString("context|assistant/v1\n")
	fmt.Fprintf(&b, "sync|%s|snapshot=%s|cards=%d\n",
		scrub(in.SyncStatus), scrub(in.SnapshotID), in.TotalCards)
	if in.Question != "" {
		question := in.Question
		if len(question) > maxQuestionLen {
			question = question[:maxQuestionLen]
		}
		fmt.Fprintf(&b, "question|%s\n", scrub(question))
	}
	if in.Deck != nil {
		fmt.Fprintf(&b, "deck|%s|main=%d|side=%d\n",
			scrub(in.Deck.Name), len(in.Deck.Main), len(in.Deck.Sideboard))
		for _, entry := range in.Deck.Main {
			fmt.Fprintf(&b, "card|%d|%s\n", entry.Quantity, scrub(entry.Name))
		}
	}
	if in.Ownership != nil {
		fmt.Fprintf(&b, "ownership|missing=%d|complete=%t\n",
			in.Ownership.TotalMissing, in.Ownership.Complete)
		for _, line := range in.Ownership.Lines {
			if line.Missing > 0 {
				fmt.Fprintf(&b, "gap|%d|%s\n", line.Missing, scrub(line.Name))
			}
		}
	}
	return b.String()
}
