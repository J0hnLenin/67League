package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/domain/settings"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type SettingsRepo struct {
	pool *pgxpool.Pool
}

func NewSettingsRepo(pool *pgxpool.Pool) *SettingsRepo {
	return &SettingsRepo{pool: pool}
}

func (r *SettingsRepo) Get(ctx context.Context, key string) (any, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var raw []byte
	err := q.QueryRow(ctx,
		`SELECT value FROM site_settings WHERE key = $1`, key,
	).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("get setting %q: %w", key, err)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("unmarshal setting: %w", err)
	}
	return v, nil
}

func (r *SettingsRepo) GetAll(ctx context.Context) (settings.SiteSettings, error) {
	q := QuerierFromCtx(ctx, r.pool)
	rows, err := q.Query(ctx, `SELECT key, value FROM site_settings`)
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}
	defer rows.Close()

	out := settings.SiteSettings{}
	for rows.Next() {
		var (
			key string
			raw []byte
		)
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, fmt.Errorf("unmarshal setting %q: %w", key, err)
		}
		out[key] = v
	}
	return out, rows.Err()
}

func (r *SettingsRepo) Set(ctx context.Context, key string, value any) error {
	q := QuerierFromCtx(ctx, r.pool)
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal setting: %w", err)
	}
	_, err = q.Exec(ctx, `
		INSERT INTO site_settings (key, value, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
	`, key, raw)
	if err != nil {
		return fmt.Errorf("set setting %q: %w", key, err)
	}
	return nil
}

func (r *SettingsRepo) SetMany(ctx context.Context, values settings.SiteSettings) error {
	// Простая реализация через последовательные Set.
	// В реальном проекте — batch через pgx.Batch или один INSERT ... ON CONFLICT.
	for k, v := range values {
		if err := r.Set(ctx, k, v); err != nil {
			return err
		}
	}
	return nil
}
