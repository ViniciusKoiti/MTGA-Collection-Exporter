// Package devmcp contém a lógica do MCP de desenvolvimento. Este arquivo
// implementa a classificação de ambiente pré-conexão (tarefa 7.2 do
// OpenSpec add-graph-workflow-harness): produção e ambiguidade são
// negadas ANTES de qualquer conexão — na dúvida, nega.
package devmcp

import (
	"fmt"
	"strings"
)

// Ambiente é a classe atribuída ao alvo de conexão.
type Ambiente string

// Classes de ambiente.
const (
	AmbienteDev      Ambiente = "development"
	AmbienteStaging  Ambiente = "staging"
	AmbienteProducao Ambiente = "production"
	AmbienteAmbiguo  Ambiente = "ambiguous"
)

// Alvo descreve a conexão pretendida: host, DSN do banco, CN do
// certificado apresentado e o marcador explícito de ambiente.
type Alvo struct {
	Host     string
	DSN      string
	CertCN   string
	Marcador string // declaração explícita: "development" ou "staging"
}

// marcadoresProducao derrubam o alvo em qualquer campo.
var marcadoresProducao = []string{"prod", "production", "release", "live"}

// hostsLocais são reconhecidos como desenvolvimento local.
var hostsLocais = []string{"localhost", "127.0.0.1", "[::1]", ".local"}

// Classifica atribui a classe do alvo. Produção vence qualquer marcador;
// sem marcador explícito de development/staging o alvo é ambíguo.
func Classifica(alvo Alvo) Ambiente {
	campos := strings.ToLower(strings.Join(
		[]string{alvo.Host, alvo.DSN, alvo.CertCN}, "\x00"))
	for _, marcador := range marcadoresProducao {
		if strings.Contains(campos, marcador) {
			return AmbienteProducao
		}
	}
	switch strings.ToLower(alvo.Marcador) {
	case string(AmbienteDev):
		if hostLocal(alvo.Host) || alvo.Host == "" {
			return AmbienteDev
		}
		return AmbienteAmbiguo // marcador dev com host remoto não fecha
	case string(AmbienteStaging):
		if strings.Contains(campos, "staging") {
			return AmbienteStaging
		}
		return AmbienteAmbiguo // marcador staging sem evidência no alvo
	default:
		return AmbienteAmbiguo
	}
}

// AutorizaConexao libera apenas development e staging classificados.
func AutorizaConexao(alvo Alvo) error {
	switch classe := Classifica(alvo); classe {
	case AmbienteDev, AmbienteStaging:
		return nil
	default:
		return fmt.Errorf("devmcp: conexão negada: ambiente %s", classe)
	}
}

func hostLocal(host string) bool {
	minusculo := strings.ToLower(host)
	for _, local := range hostsLocais {
		if minusculo == strings.TrimPrefix(local, ".") ||
			strings.HasSuffix(minusculo, local) {
			return true
		}
	}
	return strings.HasPrefix(minusculo, "127.")
}
