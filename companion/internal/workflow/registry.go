package workflow

import "fmt"

// Registry guarda as definições compiladas e aplica os tetos duros de
// limites. É montado uma vez na composição do processo; não há registro
// dinâmico em runtime nem definição vinda de modelo.
type Registry struct {
	caps Limits
	defs map[Identity]Definition
}

// NewRegistry cria o registry com os tetos duros centrais. Tetos zerados
// assumem os defaults, que então também são o máximo permitido.
func NewRegistry(caps Limits) *Registry {
	padrao := DefaultLimits()
	if caps.MaxSteps == 0 {
		caps.MaxSteps = padrao.MaxSteps
	}
	if caps.MaxToolCalls == 0 {
		caps.MaxToolCalls = padrao.MaxToolCalls
	}
	if caps.MaxRepeatedCalls == 0 {
		caps.MaxRepeatedCalls = padrao.MaxRepeatedCalls
	}
	if caps.MaxPayloadBytes == 0 {
		caps.MaxPayloadBytes = padrao.MaxPayloadBytes
	}
	if caps.ActiveDeadline == 0 {
		caps.ActiveDeadline = padrao.ActiveDeadline
	}
	return &Registry{caps: caps, defs: make(map[Identity]Definition)}
}

// Register valida a definição por completo antes de aceitá-la; identidade
// duplicada é rejeitada para garantir unicidade de versão por kind.
func (r *Registry) Register(d Definition) error {
	if _, existe := r.defs[d.Identity]; existe {
		return fmt.Errorf("%w: %s", ErrDuplicateIdentity, d.Identity)
	}
	d.Limits = d.limitesEfetivos()
	if err := validar(d, r.caps); err != nil {
		return err
	}
	r.defs[d.Identity] = d
	return nil
}

// Get devolve a definição registrada para a identidade exata.
func (r *Registry) Get(id Identity) (Definition, error) {
	d, ok := r.defs[id]
	if !ok {
		return Definition{}, fmt.Errorf("%w: %s", ErrUnknownIdentity, id)
	}
	return d, nil
}

// Identities lista as identidades registradas (para inspeção e dev-mcp).
func (r *Registry) Identities() []Identity {
	ids := make([]Identity, 0, len(r.defs))
	for id := range r.defs {
		ids = append(ids, id)
	}
	return ids
}
