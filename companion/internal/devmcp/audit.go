package devmcp

import "time"

// AuditEntry is one prompt-free audit record (task 7.5): tool name,
// sizes, duration and refusal only — arguments are never stored.
type AuditEntry struct {
	Tool          string
	RequestBytes  int
	ResponseBytes int
	Duration      time.Duration
	Refused       bool
}

// Audit copies the bounded audit trail.
func (s *Server) Audit() []AuditEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]AuditEntry(nil), s.audit...)
}

// record appends one entry, dropping the oldest past the budget.
func (s *Server) record(limits Limits, entry AuditEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audit = append(s.audit, entry)
	if len(s.audit) > limits.MaxAuditEntries {
		s.audit = s.audit[len(s.audit)-limits.MaxAuditEntries:]
	}
}
