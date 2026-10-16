package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/application/ports"
	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type SuperTitleRepo struct {
	pool *pgxpool.Pool
}

func NewSuperTitleRepo(pool *pgxpool.Pool) *SuperTitleRepo {
	return &SuperTitleRepo{pool: pool}
}

const superTitleColumns = `id, season_id, player_id, club_id, awarded_at`

func (r *SuperTitleRepo) Create(ctx context.Context, st *dombadge.SuperTitle) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		INSERT INTO super_titles (season_id, player_id, club_id, awarded_at)
		VALUES ($1, $2, $3, $4) RETURNING id
	`
	row := q.QueryRow(ctx, sql, st.SeasonID, st.PlayerID, st.ClubID, st.AwardedAt)
	if err := row.Scan((*int64)(&st.ID)); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create super title: %w", shared.ErrAlreadyExists)
		}
		return fmt.Errorf("create super title: %w", err)
	}
	return nil
}

func (r *SuperTitleRepo) GetBySeason(
	ctx context.Context,
	seasonID shared.SeasonID,
) (*dombadge.SuperTitle, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT `+superTitleColumns+` FROM super_titles WHERE season_id = $1`,
		int64(seasonID),
	)
	return scanSuperTitle(row)
}

func (r *SuperTitleRepo) List(
	ctx context.Context,
	f ports.SuperTitleFilter,
) ([]*dombadge.SuperTitle, error) {
	q := QuerierFromCtx(ctx, r.pool)

	sql := `SELECT ` + superTitleColumns + ` FROM super_titles WHERE 1=1`
	args := []any{}
	argN := 1

	if f.PlayerID != nil {
		sql += fmt.Sprintf(` AND player_id = $%d`, argN)
		args = append(args, int64(*f.PlayerID))
		argN++
	}
	if f.ClubID != nil {
		sql += fmt.Sprintf(` AND club_id = $%d`, argN)
		args = append(args, int64(*f.ClubID))
		argN++
	}

	p := f.Pagination.Normalize()
	sql += fmt.Sprintf(` ORDER BY awarded_at DESC LIMIT $%d OFFSET $%d`, argN, argN+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list super titles: %w", err)
	}
	defer rows.Close()

	out := make([]*dombadge.SuperTitle, 0)
	for rows.Next() {
		st, err := scanSuperTitle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func scanSuperTitle(row pgx.Row) (*dombadge.SuperTitle, error) {
	var st dombadge.SuperTitle
	if err := row.Scan(
		(*int64)(&st.ID), &st.SeasonID, &st.PlayerID, &st.ClubID, &st.AwardedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan super title: %w", err)
	}
	return &st, nil
}
