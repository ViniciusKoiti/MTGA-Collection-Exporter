package workflow

import (
	"context"
	"time"
)

// EventSchema é a versão corrente do envelope de evento (tarefa 6.1).
const EventSchema = "wf-event/v1"

// Event é o envelope versionado de evidência de transição. Ele carrega
// apenas identificadores, códigos e medidas — nunca estado, entradas,
// coleções, caminhos, credenciais, prompts ou payloads de ferramenta.
type Event struct {
	Schema      string
	Run         RunID
	Step        int // 0 em eventos de desfecho do run
	Graph       Identity
	Correlation string // agrupa eventos do mesmo run
	Causation   string // evento imediatamente anterior (vazio no primeiro)
	At          time.Time
	Outcome     OutcomeCode
	DurationMS  int64
	Attrs       map[string]string // somente chaves da allowlist, valores limitados
}

// EventSink recebe eventos validados; adapters reais fazem journaling
// transacional (tarefa 6.2, futura).
type EventSink interface {
	Emit(ctx context.Context, ev Event) error
}

// AttrAllowlist enumera as únicas chaves de atributo permitidas em eventos
// e o tamanho máximo de valor. A lista cresce por revisão de privacidade,
// nunca por conveniência de debug.
var AttrAllowlist = map[string]bool{
	"status":      true, // status final do run
	"error_code":  true, // código estável de erro
	"source_kind": true, // tipo de fonte de coleção (não a fonte em si)
	"provider":    true, // nome de provedor aprovado
	"count":       true, // contagem limitada (itens processados)
}

// MaxAttrValueLen limita cada valor de atributo (tarefa 6.1: atributos
// limitados e não privados).
const MaxAttrValueLen = 64
