package centralclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Rotate swaps the presented credential for a fresh one, returned
// exactly once by the central.
func (c Client) Rotate(ctx context.Context, token string) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	if err := c.do(ctx, "/v1/installations/rotate", token, nil,
		http.StatusOK, &out); err != nil {
		return "", err
	}
	return out.Token, nil
}

// Revoke disables the presented credential immediately.
func (c Client) Revoke(ctx context.Context, token string) error {
	return c.do(ctx, "/v1/installations/revoke", token, nil,
		http.StatusNoContent, nil)
}

// RequestDeletion files a deletion gated by the deletion secret; it
// deliberately needs no token — deletion works with lost credentials.
func (c Client) RequestDeletion(ctx context.Context, installationID,
	deletionSecret string) error {
	body, err := json.Marshal(map[string]string{
		"installation_id": installationID,
		"deletion_secret": deletionSecret})
	if err != nil {
		return fmt.Errorf("centralclient: encode: %w", err)
	}
	return c.do(ctx, "/v1/installations/deletion", "",
		bytes.NewReader(body), http.StatusAccepted, nil)
}

// do posts to one contract path and decodes the expected response.
func (c Client) do(ctx context.Context, path, token string,
	body io.Reader, wantStatus int, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+path, body)
	if err != nil {
		return fmt.Errorf("centralclient: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.client().Do(req)
	if err != nil {
		return fmt.Errorf("centralclient: %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		return fmt.Errorf("centralclient: %s: status %d", path, resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("centralclient: %s: decode: %w", path, err)
	}
	return nil
}
