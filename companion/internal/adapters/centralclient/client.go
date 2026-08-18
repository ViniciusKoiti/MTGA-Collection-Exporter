// Package centralclient is the desktop client of the central
// installation API (central OpenSpec task 5.2). It speaks the
// versioned wire contract only — never a central import — and no
// flow runs without explicit telemetry opt-in.
package centralclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Credentials is what enrollment yields, shown exactly once by the
// central; the caller stores them — this package never logs them.
type Credentials struct {
	InstallationID string `json:"installation_id"`
	Token          string `json:"token"`
	DeletionSecret string `json:"deletion_secret"`
}

// Client talks to the central over its public HTTP contract.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	Consent ports.ConsentStore
}

// Enroll registers this installation, gated on explicit opt-in: with
// consent absent or revoked, no request leaves the device at all.
func (c Client) Enroll(ctx context.Context) (Credentials, error) {
	state, err := c.Consent.Current(ctx)
	if err != nil {
		return Credentials{}, fmt.Errorf("centralclient: consent: %w", err)
	}
	if !state.OptIn {
		return Credentials{}, fmt.Errorf(
			"centralclient: enrollment requires explicit opt-in")
	}
	body, err := json.Marshal(map[string]any{
		"purpose": state.Purpose, "version": state.Version})
	if err != nil {
		return Credentials{}, fmt.Errorf("centralclient: encode: %w", err)
	}
	var creds Credentials
	if err := c.do(ctx, "/v1/installations", "", bytes.NewReader(body),
		http.StatusCreated, &creds); err != nil {
		return Credentials{}, err
	}
	return creds, nil
}

func (c Client) client() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}
