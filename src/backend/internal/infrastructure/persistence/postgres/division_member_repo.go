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

type DivisionMemberRepo struct {
	pool *pgxpool.Pool
}

func NewDivisionMemberRepo(pool *pgxpool.Pool) *DivisionMemberRepo {
	return &DivisionMemberRepo{pool: pool}
}

const divisionMemberColumns = `
	id, division_id, player_id, lp, wins, losses, matches_played,
	streak_current, streak_max, streaks_broken, final_place, status,
	created_at, updated_at
`

func (r *DivisionMemberRepo) Create(ctx context.Context, m *division.DivisionMember) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		INSERT INTO division_members (
			division_id, player_id, lp, wins, losses, matches_played,
			streak_current, streak_max, streaks_broken, final_place, status,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		m.DivisionID, m.PlayerID, float64(m.LP),
		m.Wins, m.Losses, m.MatchesPlayed,
		m.StreakCurrent, m.StreakMax, m.StreaksBroken,
		m.FinalPlace, string(m.Status),
		m.CreatedAt, m.UpdatedAt,
	)
	if err := row.Scan((*int64)(&m.ID)); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create division member: %w", shared.ErrAlreadyExists)
		}
		return fmt.Errorf("create division member: %w", err)
	}
	return nil
}

func (r *DivisionMemberRepo) GetByID(ctx context.Context, id shared.DivisionMemberID) (*division.DivisionMember, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT `+divisionMemberColumns+` FROM division_members WHERE id = $1`,
		int64(id),
	)
	return scanDivisionMember(row)
}

// GetByPlayerAndSeason — членство игрока в сезоне через JOIN на divisions.
func (r *DivisionMemberRepo) GetByPlayerAndSeason(
	ctx context.Context,
	playerID shared.PlayerID,
	seasonID shared.SeasonID,
) (*division.DivisionMember, error) {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
	SELECT
		m.id, m.division_id, m.player_id, m.lp, m.wins, m.losses, m.matches_played,
		m.streak_current, m.streak_max, m.streaks_broken, m.final_place, m.status,
		m.created_at, m.updated_at
	FROM division_members m
	JOIN divisions d ON d.id = m.division_id
	WHERE m.player_id = $1 AND d.season_id = $2
	LIMIT 1
`
	row := q.QueryRow(ctx, sql, int64(playerID), int64(seasonID))
	return scanDivisionMember(row)
}

func (r *DivisionMemberRepo) Update(ctx context.Context, m *division.DivisionMember) error {
	q := QuerierFromCtx(ctx, r.pool)
	tag, err := q.Exec(ctx, `
		UPDATE division_members SET
			lp = $2, wins = $3, losses = $4, matches_played = $5,
			streak_current = $6, streak_max = $7, streaks_broken = $8,
			final_place = $9, status = $10, updated_at = $11
		WHERE id = $1`,
		int64(m.ID), float64(m.LP), m.Wins, m.Losses, m.MatchesPlayed,
		m.StreakCurrent, m.StreakMax, m.StreaksBroken,
		m.FinalPlace, string(m.Status), m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update division member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *DivisionMemberRepo) ListByDivision(
	ctx context.Context,
	divisionID shared.DivisionID,
) ([]*division.DivisionMember, error) {
	q := QuerierFromCtx(ctx, r.pool)
	rows, err := q.Query(ctx,
		`SELECT `+divisionMemberColumns+` FROM division_members WHERE division_id = $1 ORDER BY id`,
		int64(divisionID),
	)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()
	return scanDivisionMembers(rows)
}

func (r *DivisionMemberRepo) ListByPlayer(
	ctx context.Context,
	playerID shared.PlayerID,
) ([]*division.DivisionMember, error) {
	q := QuerierFromCtx(ctx, r.pool)
	rows, err := q.Query(ctx,
		`SELECT `+divisionMemberColumns+` FROM division_members WHERE player_id = $1 ORDER BY id DESC`,
		int64(playerID),
	)
	if err != nil {
		return nil, fmt.Errorf("list member history: %w", err)
	}
	defer rows.Close()
	return scanDivisionMembers(rows)
}

// ---- helpers --------------------------------------------------------------

// prefixedColumns — "m.id, m.division_id, ..." для JOIN-запросов.
func prefixedColumns(prefix, cols string) string {
	// Очень простой префиксер: ожидает columns в виде "id, division_id, ...".
	// Разбиваем по запятой и обрезаем пробелы.
	out := ""
	start := 0
	for i := 0; i <= len(cols); i++ {
		if i == len(cols) || cols[i] == ',' {
			part := trimSpace(cols[start:i])
			if part != "" {
				if out != "" {
					out += ", "
				}
				out += prefix + "." + part
			}
			start = i + 1
		}
	}
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 {
		last := s[len(s)-1]
		if last != ' ' && last != '\n' && last != '\t' {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}

func scanDivisionMember(row pgx.Row) (*division.DivisionMember, error) {
	var (
		m      division.DivisionMember
		status string
		lp     float64
	)
	err := row.Scan(
		(*int64)(&m.ID), &m.DivisionID, &m.PlayerID, &lp,
		&m.Wins, &m.Losses, &m.MatchesPlayed,
		&m.StreakCurrent, &m.StreakMax, &m.StreaksBroken,
		&m.FinalPlace, &status,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan division member: %w", err)
	}
	m.LP = shared.LP(lp)
	m.Status = division.MemberStatus(status)
	return &m, nil
}

func scanDivisionMembers(rows pgx.Rows) ([]*division.DivisionMember, error) {
	out := make([]*division.DivisionMember, 0)
	for rows.Next() {
		m, err := scanDivisionMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
