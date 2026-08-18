// Package assistant is the in-app Assistant shell (OpenSpec
// introduce-agentic-go-companion, task 6.8). It is DISABLED by default:
// the zero configuration answers nothing and never touches a model.
// Only the fake provider is accepted — production provider adapters are
// deferred until provider and data-residency decisions are recorded.
package assistant

import (
	"context"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/assistctx"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// ProviderFake is the only provider the shell accepts today.
const ProviderFake = "fake"

// Config enables the shell explicitly; the zero value keeps it off.
type Config struct {
	Enabled  bool
	Provider string
}

// Shell mediates every model interaction: context is always built by the
// redacted assistctx builder — there is no other path to the provider.
type Shell struct {
	config Config
	model  ports.Model
}

// New validates the configuration at composition time. Unknown providers
// are rejected loudly: shipping one requires recording the provider and
// data-residency decision first (design decision of the companion).
func New(config Config, model ports.Model) (*Shell, error) {
	if config.Enabled && config.Provider != ProviderFake {
		return nil, fmt.Errorf(
			"assistant: provider %q deferred until provider/data-residency decisions are recorded",
			config.Provider)
	}
	if config.Enabled && model == nil {
		return nil, fmt.Errorf("assistant: enabled shell requires a model adapter")
	}
	return &Shell{config: config, model: model}, nil
}

// Ask answers a question with the redacted context. A disabled shell
// fails with the stable assistant_disabled code and never calls the
// model — the deterministic desktop works fully without it.
func (s *Shell) Ask(ctx context.Context, input assistctx.Input) (string, error) {
	if !s.config.Enabled {
		return "", apperr.New(apperr.CodeAssistantDisabled, "assistant.ask", nil)
	}
	answer, err := s.model.Respond(ctx, assistctx.Build(input))
	if err != nil {
		return "", apperr.New(apperr.CodeInternal, "assistant.ask", err)
	}
	return answer, nil
}
