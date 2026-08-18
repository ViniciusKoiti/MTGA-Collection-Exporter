package workflow

// Opções de composição do Engine: os ports opcionais são ligados na raiz
// de composição do processo, nunca em runtime.

// WithPolicy liga a política e o store de aprovações. Sem eles, qualquer
// nó que peça um efeito recebe negação (padrão seguro de RequestEffect).
func (e *Engine) WithPolicy(policy Policy, approvals ApprovalStore) *Engine {
	e.policy = policy
	e.approvals = approvals
	return e
}

// WithEvents liga o sink de eventos; o engine sempre o envolve no
// ValidatingSink para que campo proibido jamais seja persistido.
func (e *Engine) WithEvents(sink EventSink) *Engine {
	e.events = ValidatingSink{Next: sink}
	return e
}

// WithRecovery liga o store de leases usado por Recover após restart.
func (e *Engine) WithRecovery(leases LeaseStore) *Engine {
	e.leases = leases
	return e
}
