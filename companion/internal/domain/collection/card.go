// Package collection define o domínio de coleção do companion (OpenSpec
// introduce-agentic-go-companion, tarefa 2.2): identidade tipada de carta,
// observação de fonte, snapshot imutável, diagnóstico e frescor. O domínio
// não conhece Wails, SQLite, Scryfall, MCP nem modelo — só stdlib.
package collection

import "fmt"

// ArenaID é o identificador da impressão no cliente MTGA (grpId).
type ArenaID int

// OracleID é a identidade funcional da carta (Scryfall oracle_id).
type OracleID string

// PrintingID é a identidade estável da impressão no catálogo.
type PrintingID string

// CardIdentity é a identidade tipada de uma impressão de carta. Arena e
// Oracle podem estar ausentes quando a fonte não os conhece; a impressão
// resolvida sempre tem nome e set.
type CardIdentity struct {
	Printing PrintingID
	Arena    ArenaID
	Oracle   OracleID
	Name     string
	Set      string
}

// Resolvida informa se a identidade foi normalizada pelo catálogo.
func (c CardIdentity) Resolvida() bool {
	return c.Printing != "" && c.Name != ""
}

func (c CardIdentity) String() string {
	if !c.Resolvida() {
		return fmt.Sprintf("carta-nao-resolvida(arena=%d)", c.Arena)
	}
	return fmt.Sprintf("%s [%s]", c.Name, c.Set)
}
