package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestApplyMigrationsCreatesConfigSchema(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	pending, err := PendingMigrations(ctx, db)
	if err != nil {
		t.Fatalf("pending migrations: %v", err)
	}
	if len(pending) != 1 || pending[0].Name != "0001_openbudget_config.sql" {
		t.Fatalf("pending migrations = %#v", pending)
	}

	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if err := EnsureCurrent(ctx, db); err != nil {
		t.Fatalf("ensure current: %v", err)
	}

	var tableName string
	if err := db.QueryRowContext(ctx, `
SELECT name
FROM sqlite_master
WHERE type = 'table'
  AND name = 'openbudget_config'
`).Scan(&tableName); err != nil {
		t.Fatalf("query migrated table: %v", err)
	}
	if tableName != "openbudget_config" {
		t.Fatalf("tableName = %q, want openbudget_config", tableName)
	}
}

func openMigratedTestDB(tb testing.TB) *sql.DB {
	tb.Helper()

	db := openTestDB(tb)
	if err := ApplyMigrations(context.Background(), db); err != nil {
		tb.Fatalf("apply migrations: %v", err)
	}
	return db
}

func openTestDB(tb testing.TB) *sql.DB {
	tb.Helper()

	db, err := Open(filepath.Join(tb.TempDir(), "openbudget.db"))
	if err != nil {
		tb.Fatalf("open sqlite db: %v", err)
	}
	tb.Cleanup(func() {
		_ = db.Close()
	})
	return db
}
