package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type ClubBadgeRepo struct {
	pool *pgxpool.Pool
}

func NewClubBadgeRepo(pool *pgxpool.Pool) *ClubBadgeRepo {
	return &ClubBadgeRepo{pool: pool}
}

func (r *ClubBadgeRepo) Create(ctx context.Context, cb *dombadge.ClubBadge) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		INSERT INTO club_badges (club_id, badge_id, season_id, awarded_at)
		VALUES ($1, $2, $3, $4) RETURNING id
	`
	row := q.QueryRow(ctx, sql, cb.ClubID, cb.BadgeID, cb.SeasonID, cb.AwardedAt)
	if err := row.Scan((*int64)(&cb.ID)); err != nil {
		return fmt.Errorf("create club badge: %w", err)
	}
	return nil
}

func (r *ClubBadgeRepo) ListByClub(
	ctx context.Context,
	clubID shared.ClubID,
) ([]*dombadge.ClubBadge, error) {
	q := QuerierFromCtx(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT id, club_id, badge_id, season_id, awarded_at
		FROM club_badges
		WHERE club_id = $1
		ORDER BY awarded_at DESC
	`, int64(clubID))
	if err != nil {
		return nil, fmt.Errorf("list club badges: %w", err)
	}
	defer rows.Close()

	out := make([]*dombadge.ClubBadge, 0)
	for rows.Next() {
		var cb dombadge.ClubBadge
		if err := rows.Scan(
			(*int64)(&cb.ID), &cb.ClubID, &cb.BadgeID, &cb.SeasonID, &cb.AwardedAt,
		); err != nil {
			return nil, fmt.Errorf("scan club badge: %w", err)
		}
		out = append(out, &cb)
	}
	return out, rows.Err()
}

func (r *ClubBadgeRepo) Has(
	ctx context.Context,
	clubID shared.ClubID,
	badgeID shared.BadgeID,
) (bool, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var exists bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM club_badges WHERE club_id = $1 AND badge_id = $2)`,
		int64(clubID), int64(badgeID),
	).Scan(&exists)
	return exists, err
}
