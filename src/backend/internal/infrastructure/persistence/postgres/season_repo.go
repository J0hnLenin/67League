package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/application/ports"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type SeasonRepo struct {
	pool *pgxpool.Pool
}

func NewSeasonRepo(pool *pgxpool.Pool) *SeasonRepo { return &SeasonRepo{pool: pool} }

const seasonColumns = `id, number, title, start_date, end_date, status, settings, created_at, updated_at`

func (r *SeasonRepo) Create(ctx context.Context, s *domseason.Season) error {
	q := QuerierFromCtx(ctx, r.pool)
	settingsJSON, err := json.Marshal(s.Settings)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	const sql = `INSERT INTO seasons (number, title, start_date, end_date, status, settings, created_at, updated_at)
	             VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	row := q.QueryRow(ctx, sql,
		s.Number, s.Title, s.StartDate, s.EndDate,
		string(s.Status), settingsJSON, s.CreatedAt, s.UpdatedAt,
	)
	if err := row.Scan((*int64)(&s.ID)); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create season: %w", shared.ErrAlreadyExists)
		}
		return fmt.Errorf("create season: %w", err)
	}
	return nil
}

func (r *SeasonRepo) GetByID(ctx context.Context, id shared.SeasonID) (*domseason.Season, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+seasonColumns+` FROM seasons WHERE id = $1`, int64(id))
	return scanSeason(row)
}

func (r *SeasonRepo) Update(ctx context.Context, s *domseason.Season) error {
	q := QuerierFromCtx(ctx, r.pool)
	settingsJSON, err := json.Marshal(s.Settings)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	tag, err := q.Exec(ctx, `
		UPDATE seasons SET
			number = $2, title = $3, start_date = $4, end_date = $5,
			status = $6, settings = $7, updated_at = $8
		WHERE id = $1`,
		int64(s.ID), s.Number, s.Title, s.StartDate, s.EndDate,
		string(s.Status), settingsJSON, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update season: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// GetCurrent — последний по номеру сезон, не находящийся в архиве.
func (r *SeasonRepo) GetCurrent(ctx context.Context) (*domseason.Season, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `
		SELECT `+seasonColumns+` FROM seasons
		WHERE status <> 'archived'
		ORDER BY number DESC
		LIMIT 1
	`)
	return scanSeason(row)
}

// GetLatest — последний по номеру сезон независимо от статуса.
func (r *SeasonRepo) GetLatest(ctx context.Context) (*domseason.Season, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `
		SELECT `+seasonColumns+` FROM seasons
		ORDER BY number DESC
		LIMIT 1
	`)
	return scanSeason(row)
}

func (r *SeasonRepo) List(ctx context.Context, f ports.SeasonFilter) ([]*domseason.Season, error) {
	q := QuerierFromCtx(ctx, r.pool)

	sql := `SELECT ` + seasonColumns + ` FROM seasons WHERE 1=1`
	args := []any{}
	argN := 1

	if f.Status != nil {
		sql += fmt.Sprintf(` AND status = $%d`, argN)
		args = append(args, string(*f.Status))
		argN++
	}

	p := f.Pagination.Normalize()
	sql += fmt.Sprintf(` ORDER BY number DESC LIMIT $%d OFFSET $%d`, argN, argN+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list seasons: %w", err)
	}
	defer rows.Close()

	out := make([]*domseason.Season, 0)
	for rows.Next() {
		s, err := scanSeason(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func scanSeason(row pgx.Row) (*domseason.Season, error) {
	var (
		s           domseason.Season
		status      string
		settingsRaw []byte
	)
	err := row.Scan(
		(*int64)(&s.ID), &s.Number, &s.Title, &s.StartDate, &s.EndDate,
		&status, &settingsRaw, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan season: %w", err)
	}
	s.Status = domseason.Status(status)
	if err := json.Unmarshal(settingsRaw, &s.Settings); err != nil {
		return nil, fmt.Errorf("unmarshal settings: %w", err)
	}
	return &s, nil
}
