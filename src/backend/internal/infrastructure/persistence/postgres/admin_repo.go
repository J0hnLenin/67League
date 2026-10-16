package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/domain/admin"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type AdminRepo struct {
	pool *pgxpool.Pool
}

func NewAdminRepo(pool *pgxpool.Pool) *AdminRepo { return &AdminRepo{pool: pool} }

func (r *AdminRepo) Create(ctx context.Context, a *admin.Admin) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `INSERT INTO admins (login, password_hash, created_at)
	             VALUES ($1, $2, $3) RETURNING id`
	row := q.QueryRow(ctx, sql, a.Login, a.PasswordHash, a.CreatedAt)
	if err := row.Scan((*int64)(&a.ID)); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create admin: %w", shared.ErrAlreadyExists)
		}
		return fmt.Errorf("create admin: %w", err)
	}
	return nil
}

func (r *AdminRepo) GetByID(ctx context.Context, id shared.AdminID) (*admin.Admin, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT id, login, password_hash, created_at FROM admins WHERE id = $1`,
		int64(id),
	)
	return scanAdmin(row)
}

func (r *AdminRepo) GetByLogin(ctx context.Context, login string) (*admin.Admin, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT id, login, password_hash, created_at FROM admins WHERE login = $1`,
		login,
	)
	return scanAdmin(row)
}

func (r *AdminRepo) Update(ctx context.Context, a *admin.Admin) error {
	q := QuerierFromCtx(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE admins SET login = $2, password_hash = $3 WHERE id = $1`,
		int64(a.ID), a.Login, a.PasswordHash,
	)
	if err != nil {
		return fmt.Errorf("update admin: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func scanAdmin(row pgx.Row) (*admin.Admin, error) {
	var a admin.Admin
	if err := row.Scan((*int64)(&a.ID), &a.Login, &a.PasswordHash, &a.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan admin: %w", err)
	}
	return &a, nil
}
