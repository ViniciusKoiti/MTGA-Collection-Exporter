// Package policy classifica o risco de ferramentas do companion com
// negação por padrão (boundary da tarefa 2.1; a classificação completa
// com tokens de aprovação é a tarefa 6.2). Só stdlib.
package policy

// Risk é a classe de risco estável de uma ferramenta.
type Risk string

// Classes de risco válidas.
const (
	RiskRead        Risk = "read"         // leitura sem efeito
	RiskLocalEffect Risk = "local_effect" // efeito local; exige aprovação
	RiskProhibited  Risk = "prohibited"   // nunca executa
)

// Classifier aplica negação por padrão: ferramenta fora das listas é
// proibida — não existe classe implícita.
type Classifier struct {
	leituras map[string]bool
	efeitos  map[string]bool
}

// NewClassifier monta o classificador a partir das listas explícitas.
func NewClassifier(leituras, efeitos []string) *Classifier {
	c := &Classifier{leituras: make(map[string]bool), efeitos: make(map[string]bool)}
	for _, nome := range leituras {
		c.leituras[nome] = true
	}
	for _, nome := range efeitos {
		c.efeitos[nome] = true
	}
	return c
}

// Classify devolve a classe de risco; desconhecida é proibida.
func (c *Classifier) Classify(tool string) Risk {
	switch {
	case c.leituras[tool]:
		return RiskRead
	case c.efeitos[tool]:
		return RiskLocalEffect
	default:
		return RiskProhibited
	}
}
