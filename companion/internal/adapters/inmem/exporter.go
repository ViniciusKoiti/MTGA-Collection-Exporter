package inmem

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Exporter projeta snapshots nos formatos de compatibilidade em memória;
// a paridade byte-semântica com o export Python é assunto da tarefa 3.5
// do companion — este adapter cumpre o contrato do port para o harness.
type Exporter struct{}

var _ ports.Exporter = Exporter{}

// Export produz a projeção pedida sem tocar o disco.
func (Exporter) Export(
	_ context.Context,
	snap collection.Snapshot,
	formato ports.ExportFormat,
) ([]byte, error) {
	switch formato {
	case ports.ExportJSON:
		linhas := make(map[string]int, len(snap.Entries))
		for _, e := range snap.Entries {
			if !e.Unresolved {
				linhas[fmt.Sprintf("%d", e.Identity.Arena)] = e.Quantity
			}
		}
		return json.Marshal(linhas)
	case ports.ExportCSV:
		var b strings.Builder
		b.WriteString("arena_id,name,set,quantity\n")
		for _, e := range snap.Entries {
			if !e.Unresolved {
				fmt.Fprintf(&b, "%d,%s,%s,%d\n",
					e.Identity.Arena, e.Identity.Name, e.Identity.Set, e.Quantity)
			}
		}
		return []byte(b.String()), nil
	case ports.ExportText:
		var b strings.Builder
		for _, e := range snap.Entries {
			if !e.Unresolved {
				fmt.Fprintf(&b, "%d %s\n", e.Quantity, e.Identity.Name)
			}
		}
		return []byte(b.String()), nil
	default:
		return nil, fmt.Errorf("inmem: formato de export desconhecido %q", string(formato))
	}
}
