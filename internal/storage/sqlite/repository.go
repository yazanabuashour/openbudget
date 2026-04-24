package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Repository struct {
	db *sql.DB
}

type ConfigValue struct {
	Key       string
	ValueJSON string
	UpdatedAt time.Time
}

type UpsertConfigValueParams struct {
	Key       string
	ValueJSON string
	UpdatedAt time.Time
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetConfigValue(ctx context.Context, key string) (*ConfigValue, error) {
	if err := validateConfigKey(key); err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, `
SELECT key, value_json, updated_at
FROM openbudget_config
WHERE key = ?
`, key)
	value, err := scanConfigValue(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenBudget config value: %w", err)
	}
	return &value, nil
}

func (r *Repository) ListConfigValues(ctx context.Context) ([]ConfigValue, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT key, value_json, updated_at
FROM openbudget_config
ORDER BY key ASC
`)
	if err != nil {
		return nil, fmt.Errorf("failed to list OpenBudget config values: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	values := []ConfigValue{}
	for rows.Next() {
		value, err := scanConfigValue(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to list OpenBudget config values: %w", err)
	}
	return values, nil
}

func (r *Repository) UpsertConfigValue(ctx context.Context, params UpsertConfigValueParams) (ConfigValue, error) {
	if err := validateConfigKey(params.Key); err != nil {
		return ConfigValue{}, err
	}
	if !json.Valid([]byte(params.ValueJSON)) {
		return ConfigValue{}, fmt.Errorf("config value_json for %q must be valid JSON", params.Key)
	}

	row := r.db.QueryRowContext(ctx, `
INSERT INTO openbudget_config (
  key,
  value_json,
  updated_at
) VALUES (
  ?,
  ?,
  ?
)
ON CONFLICT(key) DO UPDATE SET
  value_json = excluded.value_json,
  updated_at = excluded.updated_at
RETURNING key, value_json, updated_at
`, params.Key, params.ValueJSON, serializeInstant(params.UpdatedAt))
	value, err := scanConfigValue(row)
	if err != nil {
		return ConfigValue{}, fmt.Errorf("failed to upsert OpenBudget config value: %w", err)
	}
	return value, nil
}

func (r *Repository) DeleteConfigValue(ctx context.Context, key string) (bool, error) {
	if err := validateConfigKey(key); err != nil {
		return false, err
	}

	result, err := r.db.ExecContext(ctx, `
DELETE FROM openbudget_config
WHERE key = ?
`, key)
	if err != nil {
		return false, fmt.Errorf("failed to delete OpenBudget config value: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to delete OpenBudget config value: %w", err)
	}
	return deleted > 0, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanConfigValue(row rowScanner) (ConfigValue, error) {
	var value ConfigValue
	var updatedAt string
	if err := row.Scan(&value.Key, &value.ValueJSON, &updatedAt); err != nil {
		return ConfigValue{}, err
	}

	parsedUpdatedAt, err := parseInstant(updatedAt)
	if err != nil {
		return ConfigValue{}, err
	}
	value.UpdatedAt = parsedUpdatedAt
	return value, nil
}

func validateConfigKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("config key is required")
	}
	return nil
}

func serializeInstant(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseInstant(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse instant %q: %w", value, err)
	}
	return parsed, nil
}
