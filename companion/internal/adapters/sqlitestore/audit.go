package sqlitestore

import (
	"context"
	"fmt"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// The Store also persists the redacted assistant audit trail (OpenSpec
// introduce-agentic-go-companion, task 6.3).
var _ ports.AuditLog = (*Store)(nil)

// maxAuditField bounds every stored audit field; audit rows carry codes
// and hashes, never payloads.
const maxAuditField = 128

// Append stores the record after enforcing redaction: payload-shaped or
// path-shaped values are rejected loudly instead of silently trimmed.
func (s *Store) Append(ctx context.Context, rec approvals.AuditRecord) error {
	for name, value := range map[string]string{
		"correlation": rec.Correlation, "tool": rec.Tool, "args_hash": rec.ArgsHash,
	} {
		if len(value) > maxAuditField || strings.ContainsAny(value, "{}\\/ \n") {
			return fmt.Errorf("sqlitestore: audit field %s is not redacted", name)
		}
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO audit
		(correlation, tool, args_hash, outcome, at) VALUES (?, ?, ?, ?, ?)`,
		rec.Correlation, rec.Tool, rec.ArgsHash, string(rec.Outcome),
		formataInstante(rec.At))
	return err
}

// Recent returns the newest records, oldest first within the window.
func (s *Store) Recent(ctx context.Context, limit int) ([]approvals.AuditRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT correlation, tool, args_hash,
		outcome, at FROM audit ORDER BY seq DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var records []approvals.AuditRecord
	for rows.Next() {
		var rec approvals.AuditRecord
		var outcome, at string
		if err := rows.Scan(&rec.Correlation, &rec.Tool, &rec.ArgsHash,
			&outcome, &at); err != nil {
			return nil, err
		}
		rec.Outcome = approvals.AuditOutcome(outcome)
		if rec.At, err = parseInstante(at); err != nil {
			return nil, err
		}
		records = append([]approvals.AuditRecord{rec}, records...)
	}
	return records, rows.Err()
}
