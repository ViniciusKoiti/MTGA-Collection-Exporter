package collection

// Severity classifica um diagnóstico de normalização ou de fonte.
type Severity string

// Severidades válidas.
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// Diagnostic é um registro estável de algo notável durante observação ou
// normalização. Code é estável para automação; Detail é seguro para
// apresentação e nunca carrega caminhos absolutos ou credenciais.
type Diagnostic struct {
	Code     string
	Detail   string
	Severity Severity
}

// Freshness descreve a idade do snapshot em relação à fonte.
type Freshness string

// Estados de frescor válidos.
const (
	FreshnessFresh   Freshness = "fresh"
	FreshnessStale   Freshness = "stale"
	FreshnessUnknown Freshness = "unknown"
)
