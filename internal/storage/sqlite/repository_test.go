package sqlite

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRepositoryConfigLifecycle(t *testing.T) {
	t.Parallel()

	db := openMigratedTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	updatedAt := time.Date(2026, 4, 24, 12, 30, 0, 0, time.UTC)

	missing, err := repo.GetConfigValue(ctx, "runtime.default_account")
	if err != nil {
		t.Fatalf("get missing config value: %v", err)
	}
	if missing != nil {
		t.Fatalf("missing config value = %#v, want nil", missing)
	}

	created, err := repo.UpsertConfigValue(ctx, UpsertConfigValueParams{
		Key:       "runtime.default_account",
		ValueJSON: `{"value":"checking"}`,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("upsert config value: %v", err)
	}
	if created.Key != "runtime.default_account" || created.ValueJSON != `{"value":"checking"}` || !created.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("created config value = %#v", created)
	}

	reloaded, err := repo.GetConfigValue(ctx, "runtime.default_account")
	if err != nil {
		t.Fatalf("get config value: %v", err)
	}
	if reloaded == nil || reloaded.ValueJSON != `{"value":"checking"}` {
		t.Fatalf("reloaded config value = %#v", reloaded)
	}

	updated, err := repo.UpsertConfigValue(ctx, UpsertConfigValueParams{
		Key:       "runtime.default_account",
		ValueJSON: `{"value":"savings"}`,
		UpdatedAt: updatedAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update config value: %v", err)
	}
	if updated.ValueJSON != `{"value":"savings"}` || !updated.UpdatedAt.Equal(updatedAt.Add(time.Hour)) {
		t.Fatalf("updated config value = %#v", updated)
	}

	values, err := repo.ListConfigValues(ctx)
	if err != nil {
		t.Fatalf("list config values: %v", err)
	}
	if len(values) != 1 || values[0].Key != "runtime.default_account" || values[0].ValueJSON != `{"value":"savings"}` {
		t.Fatalf("config values = %#v", values)
	}

	deleted, err := repo.DeleteConfigValue(ctx, "runtime.default_account")
	if err != nil {
		t.Fatalf("delete config value: %v", err)
	}
	if !deleted {
		t.Fatal("delete config value returned false, want true")
	}
	deleted, err = repo.DeleteConfigValue(ctx, "runtime.default_account")
	if err != nil {
		t.Fatalf("delete missing config value: %v", err)
	}
	if deleted {
		t.Fatal("delete missing config value returned true, want false")
	}
}

func TestRepositoryConfigValidation(t *testing.T) {
	t.Parallel()

	db := openMigratedTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	_, err := repo.UpsertConfigValue(ctx, UpsertConfigValueParams{
		Key:       "runtime.default_account",
		ValueJSON: `{"value":`,
		UpdatedAt: time.Date(2026, 4, 24, 12, 30, 0, 0, time.UTC),
	})
	if err == nil || !strings.Contains(err.Error(), "valid JSON") {
		t.Fatalf("invalid JSON error = %v, want valid JSON rejection", err)
	}

	_, err = repo.GetConfigValue(ctx, " ")
	if err == nil || !strings.Contains(err.Error(), "config key is required") {
		t.Fatalf("empty key error = %v, want key rejection", err)
	}
}
