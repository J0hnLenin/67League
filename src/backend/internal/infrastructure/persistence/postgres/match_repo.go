package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/application/ports"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type MatchRepo struct {
	pool *pgxpool.Pool
}

func NewMatchRepo(pool *pgxpool.Pool) *MatchRepo { return &MatchRepo{pool: pool} }

const matchColumns = `
	id, season_id, division_id, black_player_id, white_player_id, created_by,
	status, server, time_control, handicap, komi,
	winner_id, score_diff, sgf_url, link,
	result_submitted_by, confirmed_by, cancel_reason,
	created_at, accepted_at, started_at, finished_at, confirmed_at, cancelled_at
`

func (r *MatchRepo) Create(ctx context.Context, m *dommatch.Match) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		INSERT INTO matches (
			season_id, division_id, black_player_id, white_player_id, created_by,
			status, server, time_control, handicap, komi,
			winner_id, score_diff, sgf_url, link,
			result_submitted_by, confirmed_by, cancel_reason,
			created_at, accepted_at, started_at, finished_at, confirmed_at, cancelled_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		          $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		m.SeasonID, m.DivisionID, m.BlackPlayerID, m.WhitePlayerID, m.CreatedBy,
		string(m.Status), string(m.Server), m.TimeControl, m.Handicap, m.Komi,
		m.WinnerID, m.ScoreDiff, m.SGFURL, m.ExternalLink,
		m.ResultSubmittedBy, m.ConfirmedBy, m.CancelReason,
		m.CreatedAt, m.AcceptedAt, m.StartedAt, m.FinishedAt, m.ConfirmedAt, m.CancelledAt,
	)
	if err := row.Scan((*int64)(&m.ID)); err != nil {
		return fmt.Errorf("create match: %w", err)
	}
	return nil
}

func (r *MatchRepo) GetByID(ctx context.Context, id shared.MatchID) (*dommatch.Match, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+matchColumns+` FROM matches WHERE id = $1`, int64(id))
	return scanMatch(row)
}

func (r *MatchRepo) Update(ctx context.Context, m *dommatch.Match) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		UPDATE matches SET
			status = $2, server = $3, time_control = $4, handicap = $5, komi = $6,
			winner_id = $7, score_diff = $8, sgf_url = $9, link = $10,
			result_submitted_by = $11, confirmed_by = $12, cancel_reason = $13,
			accepted_at = $14, started_at = $15, finished_at = $16,
			confirmed_at = $17, cancelled_at = $18
		WHERE id = $1
	`
	tag, err := q.Exec(ctx, sql,
		int64(m.ID), string(m.Status), string(m.Server), m.TimeControl, m.Handicap, m.Komi,
		m.WinnerID, m.ScoreDiff, m.SGFURL, m.ExternalLink,
		m.ResultSubmittedBy, m.ConfirmedBy, m.CancelReason,
		m.AcceptedAt, m.StartedAt, m.FinishedAt, m.ConfirmedAt, m.CancelledAt,
	)
	if err != nil {
		return fmt.Errorf("update match: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *MatchRepo) ListByPlayer(
	ctx context.Context,
	playerID shared.PlayerID,
	f ports.MatchFilter,
) ([]*dommatch.Match, error) {
	q := QuerierFromCtx(ctx, r.pool)

	sql := `SELECT ` + matchColumns + ` FROM matches
	        WHERE (black_player_id = $1 OR white_player_id = $1)`
	args := []any{int64(playerID)}
	argN := 2

	if f.Status != nil {
		sql += fmt.Sprintf(` AND status = $%d`, argN)
		args = append(args, string(*f.Status))
		argN++
	}
	if f.Server != nil {
		sql += fmt.Sprintf(` AND server = $%d`, argN)
		args = append(args, string(*f.Server))
		argN++
	}

	p := f.Pagination.Normalize()
	sql += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argN, argN+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list matches by player: %w", err)
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) ListByDivision(
	ctx context.Context,
	divisionID shared.DivisionID,
	f ports.MatchFilter,
) ([]*dommatch.Match, error) {
	q := QuerierFromCtx(ctx, r.pool)

	sql := `SELECT ` + matchColumns + ` FROM matches WHERE division_id = $1`
	args := []any{int64(divisionID)}
	argN := 2

	if f.Status != nil {
		sql += fmt.Sprintf(` AND status = $%d`, argN)
		args = append(args, string(*f.Status))
		argN++
	}
	if f.Server != nil {
		sql += fmt.Sprintf(` AND server = $%d`, argN)
		args = append(args, string(*f.Server))
		argN++
	}

	p := f.Pagination.Normalize()
	sql += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argN, argN+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list matches by division: %w", err)
	}
	defer rows.Close()
	return scanMatches(rows)
}

// CountBetweenPlayersInSeason — только confirmed.
func (r *MatchRepo) CountBetweenPlayersInSeason(
	ctx context.Context,
	seasonID shared.SeasonID,
	p1, p2 shared.PlayerID,
) (int, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var n int
	err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM matches
		WHERE season_id = $1
		  AND status = 'confirmed'
		  AND (
			(black_player_id = $2 AND white_player_id = $3)
			OR (black_player_id = $3 AND white_player_id = $2)
		  )
	`, int64(seasonID), int64(p1), int64(p2)).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count matches between players: %w", err)
	}
	return n, nil
}

// CountConfirmedByPlayerInSeason — для бонуса «сверх t».
func (r *MatchRepo) CountConfirmedByPlayerInSeason(
	ctx context.Context,
	seasonID shared.SeasonID,
	playerID shared.PlayerID,
) (int, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var n int
	err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM matches
		WHERE season_id = $1
		  AND status = 'confirmed'
		  AND (black_player_id = $2 OR white_player_id = $2)
	`, int64(seasonID), int64(playerID)).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count player matches in season: %w", err)
	}
	return n, nil
}

// CountByPlayerAndOpponentPerServer — на будущее (аналитика).
func (r *MatchRepo) CountByPlayerAndOpponentPerServer(
	ctx context.Context,
	seasonID shared.SeasonID,
	p1, p2 shared.PlayerID,
	server dommatch.Server,
) (int, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var n int
	err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM matches
		WHERE season_id = $1
		  AND server = $4
		  AND (
			(black_player_id = $2 AND white_player_id = $3)
			OR (black_player_id = $3 AND white_player_id = $2)
		  )
	`, int64(seasonID), int64(p1), int64(p2), string(server)).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count matches per server: %w", err)
	}
	return n, nil
}

func scanMatch(row pgx.Row) (*dommatch.Match, error) {
	var (
		m              dommatch.Match
		status, server string
	)
	err := row.Scan(
		(*int64)(&m.ID), &m.SeasonID, &m.DivisionID,
		&m.BlackPlayerID, &m.WhitePlayerID, &m.CreatedBy,
		&status, &server, &m.TimeControl, &m.Handicap, &m.Komi,
		&m.WinnerID, &m.ScoreDiff, &m.SGFURL, &m.ExternalLink,
		&m.ResultSubmittedBy, &m.ConfirmedBy, &m.CancelReason,
		&m.CreatedAt, &m.AcceptedAt, &m.StartedAt, &m.FinishedAt,
		&m.ConfirmedAt, &m.CancelledAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan match: %w", err)
	}
	m.Status = dommatch.Status(status)
	m.Server = dommatch.Server(server)
	return &m, nil
}

func scanMatches(rows pgx.Rows) ([]*dommatch.Match, error) {
	out := make([]*dommatch.Match, 0)
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
