package localruntime

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yazanabuashour/openbudget/internal/storage/sqlite"
)

func TestResolvePathsUsesDatabasePathOnly(t *testing.T) {
	t.Parallel()

	databasePath := filepath.Join(t.TempDir(), "nested", "openbudget.db")
	paths, err := ResolvePaths(Config{DatabasePath: databasePath})
	if err != nil {
		t.Fatalf("resolve paths: %v", err)
	}
	if paths.DataDir != filepath.Dir(databasePath) {
		t.Fatalf("DataDir = %q, want %q", paths.DataDir, filepath.Dir(databasePath))
	}
	if paths.DatabasePath != databasePath {
		t.Fatalf("DatabasePath = %q, want %q", paths.DatabasePath, databasePath)
	}
}

func TestOpenCreatesDatabaseAndAppliesMigrations(t *testing.T) {
	t.Parallel()

	databasePath := filepath.Join(t.TempDir(), "nested", "openbudget.db")
	session, err := Open(Config{
		DatabasePath: databasePath,
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatalf("open local runtime: %v", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			t.Fatalf("close local runtime: %v", err)
		}
	}()

	if session.Paths.DatabasePath != databasePath {
		t.Fatalf("DatabasePath = %q, want %q", session.Paths.DatabasePath, databasePath)
	}
	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("stat database path: %v", err)
	}

	_, err = session.Repository.UpsertConfigValue(context.Background(), sqlite.UpsertConfigValueParams{
		Key:       "runtime.default_account",
		ValueJSON: `{"value":"checking"}`,
		UpdatedAt: time.Date(2026, 4, 24, 12, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("upsert config through runtime repository: %v", err)
	}
}
