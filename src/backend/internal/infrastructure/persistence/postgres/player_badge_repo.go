package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type PlayerBadgeRepo struct {
	pool *pgxpool.Pool
}

func NewPlayerBadgeRepo(pool *pgxpool.Pool) *PlayerBadgeRepo {
	return &PlayerBadgeRepo{pool: pool}
}

func (r *PlayerBadgeRepo) Create(ctx context.Context, pb *dombadge.PlayerBadge) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		INSERT INTO player_badges (player_id, badge_id, season_id, awarded_at, is_active)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		pb.PlayerID, pb.BadgeID, pb.SeasonID, pb.AwardedAt, pb.IsActive,
	)
	if err := row.Scan((*int64)(&pb.ID)); err != nil {
		return fmt.Errorf("create player badge: %w", err)
	}
	return nil
}

func (r *PlayerBadgeRepo) ListByPlayer(
	ctx context.Context,
	playerID shared.PlayerID,
) ([]*dombadge.PlayerBadge, error) {
	q := QuerierFromCtx(ctx, r.pool)
	return r.list(ctx, q, `
		SELECT id, player_id, badge_id, season_id, awarded_at, is_active
		FROM player_badges
		WHERE player_id = $1
		ORDER BY awarded_at DESC
	`, int64(playerID))
}

func (r *PlayerBadgeRepo) ListByPlayerAndSeason(
	ctx context.Context,
	playerID shared.PlayerID,
	seasonID shared.SeasonID,
) ([]*dombadge.PlayerBadge, error) {
	q := QuerierFromCtx(ctx, r.pool)
	return r.list(ctx, q, `
		SELECT id, player_id, badge_id, season_id, awarded_at, is_active
		FROM player_badges
		WHERE player_id = $1 AND season_id = $2
		ORDER BY awarded_at DESC
	`, int64(playerID), int64(seasonID))
}

func (r *PlayerBadgeRepo) Has(
	ctx context.Context,
	playerID shared.PlayerID,
	badgeID shared.BadgeID,
) (bool, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var exists bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM player_badges WHERE player_id = $1 AND badge_id = $2)`,
		int64(playerID), int64(badgeID),
	).Scan(&exists)
	return exists, err
}

func (r *PlayerBadgeRepo) DeactivateSeasonal(ctx context.Context, playerID shared.PlayerID) error {
	q := QuerierFromCtx(ctx, r.pool)
	_, err := q.Exec(ctx,
		`UPDATE player_badges SET is_active = false WHERE player_id = $1 AND season_id IS NOT NULL`,
		int64(playerID),
	)
	if err != nil {
		return fmt.Errorf("deactivate seasonal: %w", err)
	}
	return nil
}

func (r *PlayerBadgeRepo) list(
	ctx context.Context,
	q Querier,
	sql string,
	args ...any,
) ([]*dombadge.PlayerBadge, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list player badges: %w", err)
	}
	defer rows.Close()

	out := make([]*dombadge.PlayerBadge, 0)
	for rows.Next() {
		var pb dombadge.PlayerBadge
		if err := rows.Scan(
			(*int64)(&pb.ID), &pb.PlayerID, &pb.BadgeID,
			&pb.SeasonID, &pb.AwardedAt, &pb.IsActive,
		); err != nil {
			return nil, fmt.Errorf("scan player badge: %w", err)
		}
		out = append(out, &pb)
	}
	return out, rows.Err()
}
