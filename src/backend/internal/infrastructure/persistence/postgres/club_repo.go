package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/application/ports"
	domclub "github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type ClubRepo struct {
	pool *pgxpool.Pool
}

func NewClubRepo(pool *pgxpool.Pool) *ClubRepo { return &ClubRepo{pool: pool} }

const clubColumns = `id, name, city, university, logo_url, created_by, status, created_at, updated_at`

func (r *ClubRepo) Create(ctx context.Context, c *domclub.Club) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `INSERT INTO clubs (name, city, university, logo_url, created_by, status, created_at, updated_at)
	             VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	row := q.QueryRow(ctx, sql,
		c.Name, c.City, c.University, c.LogoURL, c.CreatedBy,
		string(c.Status), c.CreatedAt, c.UpdatedAt,
	)
	if err := row.Scan((*int64)(&c.ID)); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create club: %w", shared.ErrAlreadyExists)
		}
		return fmt.Errorf("create club: %w", err)
	}
	return nil
}

func (r *ClubRepo) GetByID(ctx context.Context, id shared.ClubID) (*domclub.Club, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+clubColumns+` FROM clubs WHERE id = $1`, int64(id))
	return scanClub(row)
}

func (r *ClubRepo) GetByName(ctx context.Context, name string) (*domclub.Club, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+clubColumns+` FROM clubs WHERE name = $1`, name)
	return scanClub(row)
}

func (r *ClubRepo) Update(ctx context.Context, c *domclub.Club) error {
	q := QuerierFromCtx(ctx, r.pool)
	tag, err := q.Exec(ctx, `
		UPDATE clubs SET
			name = $2, city = $3, university = $4, logo_url = $5,
			created_by = $6, status = $7, updated_at = $8
		WHERE id = $1`,
		int64(c.ID), c.Name, c.City, c.University, c.LogoURL,
		c.CreatedBy, string(c.Status), c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update club: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *ClubRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var exists bool
	err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM clubs WHERE name = $1)`, name).Scan(&exists)
	return exists, err
}

func (r *ClubRepo) List(ctx context.Context, f ports.ClubFilter) ([]*domclub.Club, error) {
	q := QuerierFromCtx(ctx, r.pool)

	sql := `SELECT ` + clubColumns + ` FROM clubs WHERE 1=1`
	args := []any{}
	argN := 1

	if f.City != nil {
		sql += fmt.Sprintf(` AND city = $%d`, argN)
		args = append(args, *f.City)
		argN++
	}
	if f.Status != nil {
		sql += fmt.Sprintf(` AND status = $%d`, argN)
		args = append(args, string(*f.Status))
		argN++
	}

	p := f.Pagination.Normalize()
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, argN, argN+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	defer rows.Close()

	out := make([]*domclub.Club, 0)
	for rows.Next() {
		c, err := scanClub(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func scanClub(row pgx.Row) (*domclub.Club, error) {
	var (
		c      domclub.Club
		status string
	)
	err := row.Scan(
		(*int64)(&c.ID), &c.Name, &c.City, &c.University, &c.LogoURL,
		&c.CreatedBy, &status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan club: %w", err)
	}
	c.Status = domclub.Status(status)
	return &c, nil
}
