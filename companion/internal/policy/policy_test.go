package policy

import "testing"

func TestClassificacaoNegaPorPadrao(t *testing.T) {
	classificador := NewClassifier(
		[]string{"query-collection", "check-deck"},
		[]string{"export-approved"},
	)
	casos := map[string]Risk{
		"query-collection": RiskRead,
		"check-deck":       RiskRead,
		"export-approved":  RiskLocalEffect,
		"memory-scan":      RiskProhibited, // nunca listada => negada
		"shell":            RiskProhibited,
		"":                 RiskProhibited,
	}
	for tool, esperado := range casos {
		if got := classificador.Classify(tool); got != esperado {
			t.Errorf("%q: esperava %s, veio %s", tool, esperado, got)
		}
	}
}
