package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type TitleRepo struct {
	pool *pgxpool.Pool
}

func NewTitleRepo(pool *pgxpool.Pool) *TitleRepo { return &TitleRepo{pool: pool} }

const titleColumns = `id, code, name, badge_id`

func (r *TitleRepo) Create(ctx context.Context, t *dombadge.Title) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `INSERT INTO titles (code, name, badge_id) VALUES ($1, $2, $3) RETURNING id`
	row := q.QueryRow(ctx, sql, t.Code, t.Name, t.BadgeID)
	if err := row.Scan((*int64)(&t.ID)); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create title: %w", shared.ErrAlreadyExists)
		}
		return fmt.Errorf("create title: %w", err)
	}
	return nil
}

func (r *TitleRepo) GetByID(ctx context.Context, id shared.TitleID) (*dombadge.Title, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+titleColumns+` FROM titles WHERE id = $1`, int64(id))
	return scanTitle(row)
}

func (r *TitleRepo) GetByCode(ctx context.Context, code string) (*dombadge.Title, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+titleColumns+` FROM titles WHERE code = $1`, code)
	return scanTitle(row)
}

func (r *TitleRepo) List(ctx context.Context) ([]*dombadge.Title, error) {
	q := QuerierFromCtx(ctx, r.pool)
	rows, err := q.Query(ctx, `SELECT `+titleColumns+` FROM titles ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list titles: %w", err)
	}
	defer rows.Close()

	out := make([]*dombadge.Title, 0)
	for rows.Next() {
		t, err := scanTitle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func scanTitle(row pgx.Row) (*dombadge.Title, error) {
	var t dombadge.Title
	if err := row.Scan((*int64)(&t.ID), &t.Code, &t.Name, &t.BadgeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan title: %w", err)
	}
	return &t, nil
}
