package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type DivisionRepo struct {
	pool *pgxpool.Pool
}

func NewDivisionRepo(pool *pgxpool.Pool) *DivisionRepo { return &DivisionRepo{pool: pool} }

const divisionColumns = `
	id, season_id, code, alias, rating_min, rating_max,
	"order", status, rotation_up, rotation_down, created_at, updated_at
`

func (r *DivisionRepo) Create(ctx context.Context, d *division.Division) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		INSERT INTO divisions (
			season_id, code, alias, rating_min, rating_max,
			"order", status, rotation_up, rotation_down, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		d.SeasonID, d.Code, d.Alias, int(d.RatingMin), int(d.RatingMax),
		d.Order, string(d.Status), d.RotationUp, d.RotationDown,
		d.CreatedAt, d.UpdatedAt,
	)
	if err := row.Scan((*int64)(&d.ID)); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create division: %w", shared.ErrAlreadyExists)
		}
		return fmt.Errorf("create division: %w", err)
	}
	return nil
}

func (r *DivisionRepo) GetByID(ctx context.Context, id shared.DivisionID) (*division.Division, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+divisionColumns+` FROM divisions WHERE id = $1`, int64(id))
	return scanDivision(row)
}

func (r *DivisionRepo) Update(ctx context.Context, d *division.Division) error {
	q := QuerierFromCtx(ctx, r.pool)
	tag, err := q.Exec(ctx, `
		UPDATE divisions SET
			code = $2, alias = $3, rating_min = $4, rating_max = $5,
			"order" = $6, status = $7, rotation_up = $8, rotation_down = $9,
			updated_at = $10
		WHERE id = $1`,
		int64(d.ID), d.Code, d.Alias, int(d.RatingMin), int(d.RatingMax),
		d.Order, string(d.Status), d.RotationUp, d.RotationDown,
		d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update division: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *DivisionRepo) ListBySeason(ctx context.Context, seasonID shared.SeasonID) ([]*division.Division, error) {
	q := QuerierFromCtx(ctx, r.pool)
	rows, err := q.Query(ctx,
		`SELECT `+divisionColumns+` FROM divisions WHERE season_id = $1 ORDER BY "order"`,
		int64(seasonID),
	)
	if err != nil {
		return nil, fmt.Errorf("list divisions: %w", err)
	}
	defer rows.Close()
	return scanDivisions(rows)
}

// FindForRating — первый дивизион (по порядку), в диапазон которого попадает рейтинг.
func (r *DivisionRepo) FindForRating(
	ctx context.Context,
	seasonID shared.SeasonID,
	rating shared.Rating,
) (*division.Division, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `
		SELECT `+divisionColumns+` FROM divisions
		WHERE season_id = $1
		  AND rating_min <= $2
		  AND (rating_max = 0 OR rating_max >= $2)
		ORDER BY "order"
		LIMIT 1
	`, int64(seasonID), int(rating))
	return scanDivision(row)
}

func (r *DivisionRepo) GetLowest(ctx context.Context, seasonID shared.SeasonID) (*division.Division, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `
		SELECT `+divisionColumns+` FROM divisions
		WHERE season_id = $1
		ORDER BY "order" DESC
		LIMIT 1
	`, int64(seasonID))
	return scanDivision(row)
}

func (r *DivisionRepo) GetHighest(ctx context.Context, seasonID shared.SeasonID) (*division.Division, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `
		SELECT `+divisionColumns+` FROM divisions
		WHERE season_id = $1
		ORDER BY "order"
		LIMIT 1
	`, int64(seasonID))
	return scanDivision(row)
}

func scanDivision(row pgx.Row) (*division.Division, error) {
	var (
		d         division.Division
		status    string
		ratingMin int
		ratingMax int
	)
	err := row.Scan(
		(*int64)(&d.ID), &d.SeasonID, &d.Code, &d.Alias,
		&ratingMin, &ratingMax, &d.Order, &status,
		&d.RotationUp, &d.RotationDown,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan division: %w", err)
	}
	d.Status = division.Status(status)
	d.RatingMin = shared.Rating(ratingMin)
	d.RatingMax = shared.Rating(ratingMax)
	return &d, nil
}

func scanDivisions(rows pgx.Rows) ([]*division.Division, error) {
	out := make([]*division.Division, 0)
	for rows.Next() {
		d, err := scanDivision(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
