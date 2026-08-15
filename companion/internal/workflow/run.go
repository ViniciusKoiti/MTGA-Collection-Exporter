package workflow

import "time"

// RunID identifica uma execução de grafo; vem sempre do port de IDs,
// nunca de geração implícita, para manter o harness determinístico.
type RunID string

// RunStatus é o estado persistido de um run.
type RunStatus string

// Estados válidos de um run. Waiting não consome deadline ativo.
const (
	RunPending   RunStatus = "pending"
	RunActive    RunStatus = "active"
	RunWaiting   RunStatus = "waiting_approval"
	RunSucceeded RunStatus = "succeeded"
	RunFailed    RunStatus = "failed"
	RunCancelled RunStatus = "cancelled"
)

// Terminal informa se o run chegou a um desfecho definitivo.
func (s RunStatus) Terminal() bool {
	return s == RunSucceeded || s == RunFailed || s == RunCancelled
}

// Run é o checkpoint autoritativo de uma execução: fixa a identidade do
// grafo, o nó corrente, o estado tipado e uma versão otimista que protege
// os commits transacionais contra escritores concorrentes.
type Run struct {
	ID        RunID
	Graph     Identity
	Status    RunStatus
	Current   NodeID
	State     State
	Outcome   OutcomeCode // preenchido apenas em status terminal
	Version   int         // versão otimista do checkpoint
	StartedAt time.Time
	UpdatedAt time.Time
}

// Step registra uma tentativa de execução de um nó, para evidência e replay.
// Não carrega payloads: entradas e saídas ficam no checkpoint, e eventos
// são sempre redigidos (spec workflow-observability).
type Step struct {
	Run      RunID
	Index    int
	Node     NodeID
	Attempt  int
	Outcome  OutcomeCode
	Err      string // código estável do erro tipado, vazio em sucesso
	Started  time.Time
	Finished time.Time
}
