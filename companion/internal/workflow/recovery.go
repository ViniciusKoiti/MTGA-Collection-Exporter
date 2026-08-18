package workflow

import (
	"context"
	"errors"
	"time"
)

// LeaseStore controla a posse de runs recuperáveis: um único dono por vez,
// com prazo limitado; lease expirado pode ser tomado por outro processo.
type LeaseStore interface {
	// AcquireLease toma o run se ele estiver sem dono, com lease expirado
	// em relação a `agora`, ou já pertencer ao mesmo dono.
	AcquireLease(ctx context.Context, id RunID, owner string, agora, until time.Time) (bool, error)
	// ReleaseLease devolve o run; ignora silenciosamente dono divergente.
	ReleaseLease(ctx context.Context, id RunID, owner string) error
	// RecoverableRuns lista runs ativos sem lease vigente em `agora`.
	RecoverableRuns(ctx context.Context, agora time.Time) ([]RunID, error)
	// AbandonExpired encerra como abandonado todo run não terminal sem
	// lease vigente cujo último checkpoint é anterior a `cutoff`.
	AbandonExpired(ctx context.Context, agora, cutoff time.Time) (int, error)
}

// ErrLeaseHeld indica run já possuído por outro processo vivo.
var ErrLeaseHeld = errors.New("workflow: lease do run pertence a outro dono")

// Outcomes estáveis do fluxo de recuperação.
const (
	OutcomeIncompatibleVersion OutcomeCode = "incompatible_graph_version"
	OutcomeRecoveryRefused     OutcomeCode = "recovery_refused"
	OutcomeAbandoned           OutcomeCode = "abandoned"
)
