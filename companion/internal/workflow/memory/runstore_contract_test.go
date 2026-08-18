package memory_test

import (
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/storetest"
)

// O adapter in-memory passa na mesma suíte de contrato que o SQLite.
func TestRunStoreInMemoryCumpreContrato(t *testing.T) {
	storetest.Executa(t, func(t *testing.T) wf.RunStore {
		return memory.NewRunStore()
	})
}
