package postgres

import (
	"context"
	"errors"
	"testing"
)

// TestCredentialRotationIsCompareAndSwap: only the current hash can
// rotate, and a revoked installation cannot rotate at all.
func TestCredentialRotationIsCompareAndSwap(t *testing.T) {
	pool := repoPool(t)
	repo := InstallationsRepo{Q: pool}
	ctx := context.Background()
	inst := Installation{ID: "rot-1", TokenHash: "h-old", DeletionHash: "d-1"}
	if err := repo.Enroll(ctx, inst); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	if err := repo.RotateToken(ctx, "rot-1", "h-stale", "h-new"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("stale hash must not rotate: %v", err)
	}
	if err := repo.RotateToken(ctx, "rot-1", "h-old", "h-new"); err != nil {
		t.Fatalf("current hash must rotate: %v", err)
	}
	if err := repo.Revoke(ctx, "rot-1"); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if err := repo.RotateToken(ctx, "rot-1", "h-new", "h-3"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("revoked installation must not rotate: %v", err)
	}
}

// TestDeletionRequiresTheDeletionSecret: a wrong secret files nothing
// and looks exactly like an unknown installation; the right secret
// files the request and revokes atomically.
func TestDeletionRequiresTheDeletionSecret(t *testing.T) {
	pool := repoPool(t)
	repo := InstallationsRepo{Q: pool}
	ctx := context.Background()
	inst := Installation{ID: "del-1", TokenHash: "h-1", DeletionHash: "d-secret"}
	if err := repo.Enroll(ctx, inst); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	if err := RequestDeletion(ctx, pool, "del-1", "d-wrong"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("wrong secret must map to not-found: %v", err)
	}
	var filed int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM deletion_requests
		WHERE installation_id = 'del-1'`).Scan(&filed); err != nil || filed != 0 {
		t.Fatalf("wrong secret must file nothing: %d %v", filed, err)
	}
	if err := RequestDeletion(ctx, pool, "del-1", "d-secret"); err != nil {
		t.Fatalf("right secret must file the deletion: %v", err)
	}
	loaded, err := repo.ByID(ctx, "del-1")
	if err != nil || !loaded.Revoked {
		t.Fatalf("deletion must revoke the installation: %+v %v", loaded, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM deletion_requests
		WHERE installation_id = 'del-1'`).Scan(&filed); err != nil || filed != 1 {
		t.Fatalf("deletion request must be filed once: %d %v", filed, err)
	}
}
