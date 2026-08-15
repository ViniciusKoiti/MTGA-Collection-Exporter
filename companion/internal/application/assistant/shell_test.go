package assistant

import (
	"context"
	"strings"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/assistctx"
)

// capturingModel records the context it receives so tests can prove the
// redaction happened before anything reached the provider boundary.
type capturingModel struct {
	received string
}

func (m *capturingModel) Respond(_ context.Context, contextText string) (string, error) {
	m.received = contextText
	return "answer", nil
}

func TestShellIsDisabledByDefaultAndNeverCallsTheModel(t *testing.T) {
	model := &capturingModel{}
	shell, err := New(Config{}, model) // zero config: disabled
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	_, err = shell.Ask(t.Context(), assistctx.Input{Question: "help"})
	if apperr.CodeOf(err) != apperr.CodeAssistantDisabled {
		t.Fatalf("disabled shell must fail with the stable code: %v", err)
	}
	if model.received != "" {
		t.Fatal("disabled shell must never call the model")
	}
}

func TestEnabledFakeShellReceivesOnlyRedactedContext(t *testing.T) {
	model := &capturingModel{}
	shell, err := New(Config{Enabled: true, Provider: ProviderFake}, model)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	answer, err := shell.Ask(t.Context(), assistctx.Input{
		SyncStatus: "fresh", SnapshotID: "snap-1", TotalCards: 42,
		Question: `check C:\Users\me\player.log with password=hunter2`,
	})
	if err != nil || answer != "answer" {
		t.Fatalf("ask: %q (%v)", answer, err)
	}
	for _, leaked := range []string{"C:", "player.log", "hunter2"} {
		if strings.Contains(model.received, leaked) {
			t.Fatalf("model received unredacted content %q:\n%s", leaked, model.received)
		}
	}
	if !strings.Contains(model.received, "context|assistant/v1") {
		t.Fatalf("model should receive the versioned context:\n%s", model.received)
	}
}

func TestProductionProvidersAreDeferred(t *testing.T) {
	if _, err := New(Config{Enabled: true, Provider: "openai"}, inmem.FakeModel{}); err == nil {
		t.Fatal("unknown providers must be rejected until decisions are recorded")
	}
	if _, err := New(Config{Enabled: true, Provider: ProviderFake}, nil); err == nil {
		t.Fatal("enabled shell without a model adapter must fail at composition")
	}
	if _, err := New(Config{Enabled: true, Provider: ProviderFake}, inmem.FakeModel{}); err != nil {
		t.Fatalf("fake provider is the supported one: %v", err)
	}
}
