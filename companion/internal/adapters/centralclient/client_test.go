package centralclient

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

type consent struct{ state ports.ConsentState }

func (c consent) Current(context.Context) (ports.ConsentState, error) {
	return c.state, nil
}

func TestOptOutMeansNoRequestEverLeaves(t *testing.T) {
	var requests atomic.Int64
	server := stubCentral(t, &requests)
	client := Client{BaseURL: server.URL, Consent: consent{}}
	if _, err := client.Enroll(context.Background()); err == nil {
		t.Fatal("enrollment without opt-in must be refused")
	}
	if requests.Load() != 0 {
		t.Fatalf("opt-out must produce zero requests, got %d", requests.Load())
	}
}

func TestEnrollRotateRevokeAndDeletionFlows(t *testing.T) {
	var requests atomic.Int64
	server := stubCentral(t, &requests)
	client := Client{BaseURL: server.URL, Consent: consent{
		state: ports.ConsentState{OptIn: true, Purpose: "product", Version: 2}}}
	ctx := context.Background()
	creds, err := client.Enroll(ctx)
	if err != nil || creds.Token != "tok-1" || creds.DeletionSecret == "" {
		t.Fatalf("enroll: %+v %v", creds, err)
	}
	fresh, err := client.Rotate(ctx, creds.Token)
	if err != nil || fresh != "tok-2" {
		t.Fatalf("rotate: %q %v", fresh, err)
	}
	if _, err := client.Rotate(ctx, "stale"); err == nil {
		t.Fatal("a stale token must not rotate")
	}
	if err := client.Revoke(ctx, fresh); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if err := client.RequestDeletion(ctx, creds.InstallationID, "wrong"); err == nil {
		t.Fatal("a wrong deletion secret must be refused")
	}
	if err := client.RequestDeletion(ctx, creds.InstallationID,
		creds.DeletionSecret); err != nil {
		t.Fatalf("deletion: %v", err)
	}
}
