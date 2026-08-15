package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestRoleGrantAndDenialMatrix proves every runtime role can do exactly
// what it needs and nothing more, and that no runtime role is superuser
// or BYPASSRLS. Skipped locally without Docker, mandatory in CI.
func TestRoleGrantAndDenialMatrix(t *testing.T) {
	ctx := context.Background()
	container, err := pgcontainer.Run(ctx, "postgres:16-alpine",
		pgcontainer.WithDatabase("central"),
		pgcontainer.WithUsername("central"),
		pgcontainer.WithPassword("central"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(90*time.Second)))
	if err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("CI must run the role suite: %v", err)
		}
		t.Skipf("docker unavailable locally; suite is validated in CI: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	seed(t, db)

	for _, tc := range grantMatrix() {
		err := runAs(ctx, db, tc.role, tc.stmt)
		if tc.allowed && err != nil {
			t.Errorf("%s should be ALLOWED %q: %v", tc.role, tc.stmt, err)
		}
		if !tc.allowed && err == nil {
			t.Errorf("%s should be DENIED %q", tc.role, tc.stmt)
		}
	}

	rows, err := db.QueryContext(ctx, `SELECT rolname, rolsuper, rolbypassrls
		FROM pg_roles WHERE rolname LIKE 'central_%'`)
	if err != nil {
		t.Fatalf("pg_roles: %v", err)
	}
	defer func() { _ = rows.Close() }()
	roles := 0
	for rows.Next() {
		var name string
		var super, bypass bool
		if err := rows.Scan(&name, &super, &bypass); err != nil {
			t.Fatalf("scan: %v", err)
		}
		roles++
		if super || bypass {
			t.Errorf("%s must never be superuser or BYPASSRLS", name)
		}
	}
	if roles != 7 {
		t.Fatalf("expected the 7 central roles, found %d", roles)
	}
}
