// Package catalog define os tipos de domínio da publicação de catálogo.
// O domínio não conhece HTTP, banco, storage nem concorrência.
package catalog

// ProviderJob identifica uma busca aprovada em um provedor externo.
type ProviderJob struct {
	Provider string // nome do provedor aprovado
	Kind     string // tipo de catálogo (cards, meta)
}

// Observation é um registro bruto obtido de um provedor, ainda não validado.
type Observation struct {
	Provider string
	GrpID    int
	Payload  map[string]string
}

// Card é uma carta normalizada e validada, pronta para persistência.
type Card struct {
	GrpID int
	Name  string
	Set   string
}

// Snapshot é o conteúdo canônico e imutável de uma publicação.
// A ordem de Cards é determinística (GrpID crescente) por construção.
type Snapshot struct {
	SchemaVersion string
	Cards         []Card
}

// ObjectRef aponta para um snapshot imutável já gravado no object storage.
type ObjectRef struct {
	Key    string
	SHA256 string
	Size   int64
}

// Manifest é o manifesto assinado que os clientes verificam antes de ativar.
type Manifest struct {
	KeyID     string
	Signature string
	Object    ObjectRef
	Schema    string
}
