package sqlitestore

import (
	"context"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
)

// ActivityEntry is one row of the Assistant activity view: an interaction
// (correlation) with its latest outcome and how many events it produced.
type ActivityEntry struct {
	Correlation string
	Tool        string
	ArgsHash    string
	Outcome     approvals.AuditOutcome
	At          time.Time
	Events      int
}

// Activity groups the trail by correlation, newest interaction first,
// each with its latest outcome — requested, approved, executed, failed
// or denied — ready for the Assistant activity screen (task 6.3).
func (s *Store) Activity(ctx context.Context, limit int) ([]ActivityEntry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT correlation, tool, args_hash,
		outcome, at, cnt FROM (
			SELECT correlation, tool, args_hash, outcome, at, seq,
			       COUNT(*) OVER (PARTITION BY correlation) AS cnt,
			       ROW_NUMBER() OVER (PARTITION BY correlation ORDER BY seq DESC) AS rn
			FROM audit)
		WHERE rn = 1 ORDER BY seq DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var entries []ActivityEntry
	for rows.Next() {
		var entry ActivityEntry
		var outcome, at string
		if err := rows.Scan(&entry.Correlation, &entry.Tool, &entry.ArgsHash,
			&outcome, &at, &entry.Events); err != nil {
			return nil, err
		}
		entry.Outcome = approvals.AuditOutcome(outcome)
		if entry.At, err = parseInstante(at); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
