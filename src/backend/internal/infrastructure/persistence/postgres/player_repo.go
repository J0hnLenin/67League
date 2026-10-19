package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/application/ports"
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type PlayerRepo struct {
	pool *pgxpool.Pool
}

func NewPlayerRepo(pool *pgxpool.Pool) *PlayerRepo { return &PlayerRepo{pool: pool} }

const playerColumns = `
	id, full_name, email, password_hash, nickname,
	kgs_nick, ogs_nick, fox_nick, rfg_id, sputnik_id,
	city, rating, club_id, stones, lp_total,
	status, active_title_id, created_at, updated_at
`

func (r *PlayerRepo) Create(ctx context.Context, p *domplayer.Player) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		INSERT INTO players (
			full_name, email, password_hash, nickname,
			kgs_nick, ogs_nick, fox_nick, rfg_id, sputnik_id,
			city, rating, club_id, stones, lp_total,
			status, active_title_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		          $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		p.FullName, string(p.Email), p.PasswordHash, string(p.Nickname),
		p.KGSNick, p.OGSNick, p.FoxNick, p.RFGID, p.SputnikID,
		p.City, int(p.Rating), p.ClubID, float64(p.Stones), float64(p.LPTotal),
		string(p.Status), p.ActiveTitleID, p.CreatedAt, p.UpdatedAt,
	)
	if err := row.Scan((*int64)(&p.ID)); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create player: %w", shared.ErrAlreadyExists)
		}
		return fmt.Errorf("create player: %w", err)
	}
	return nil
}

func (r *PlayerRepo) GetByID(ctx context.Context, id shared.PlayerID) (*domplayer.Player, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+playerColumns+` FROM players WHERE id = $1`, int64(id))
	return scanPlayer(row)
}

func (r *PlayerRepo) GetByEmail(ctx context.Context, e shared.Email) (*domplayer.Player, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+playerColumns+` FROM players WHERE email = $1`, string(e))
	return scanPlayer(row)
}

func (r *PlayerRepo) GetByNickname(ctx context.Context, n shared.Nickname) (*domplayer.Player, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+playerColumns+` FROM players WHERE nickname = $1`, string(n))
	return scanPlayer(row)
}

func (r *PlayerRepo) Update(ctx context.Context, p *domplayer.Player) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		UPDATE players SET
			full_name = $2, email = $3, password_hash = $4, nickname = $5,
			kgs_nick = $6, ogs_nick = $7, fox_nick = $8, rfg_id = $9, sputnik_id = $10,
			city = $11, rating = $12, club_id = $13, stones = $14, lp_total = $15,
			status = $16, active_title_id = $17, updated_at = $18
		WHERE id = $1
	`
	tag, err := q.Exec(ctx, sql,
		int64(p.ID), p.FullName, string(p.Email), p.PasswordHash, string(p.Nickname),
		p.KGSNick, p.OGSNick, p.FoxNick, p.RFGID, p.SputnikID,
		p.City, int(p.Rating), p.ClubID, float64(p.Stones), float64(p.LPTotal),
		string(p.Status), p.ActiveTitleID, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update player: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *PlayerRepo) ExistsByEmail(ctx context.Context, e shared.Email) (bool, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var exists bool
	err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM players WHERE email = $1)`, string(e)).Scan(&exists)
	return exists, err
}

func (r *PlayerRepo) ExistsByNickname(ctx context.Context, n shared.Nickname) (bool, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var exists bool
	err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM players WHERE nickname = $1)`, string(n)).Scan(&exists)
	return exists, err
}

func (r *PlayerRepo) ListByClub(ctx context.Context, clubID shared.ClubID) ([]*domplayer.Player, error) {
	q := QuerierFromCtx(ctx, r.pool)
	rows, err := q.Query(ctx,
		`SELECT `+playerColumns+` FROM players WHERE club_id = $1 ORDER BY id`,
		int64(clubID),
	)
	if err != nil {
		return nil, fmt.Errorf("list players by club: %w", err)
	}
	defer rows.Close()
	return scanPlayers(rows)
}

func (r *PlayerRepo) CountByClub(ctx context.Context, clubID shared.ClubID) (int, error) {
	q := QuerierFromCtx(ctx, r.pool)
	var n int
	err := q.QueryRow(ctx, `SELECT COUNT(*) FROM players WHERE club_id = $1`, int64(clubID)).Scan(&n)
	return n, err
}

func (r *PlayerRepo) List(ctx context.Context, f ports.PlayerFilter) ([]*domplayer.Player, error) {
	q := QuerierFromCtx(ctx, r.pool)

	// Простой динамический фильтр.
	sql := `SELECT ` + playerColumns + ` FROM players WHERE 1=1`
	args := []any{}
	argN := 1

	if f.Status != nil {
		sql += fmt.Sprintf(` AND status = $%d`, argN)
		args = append(args, string(*f.Status))
		argN++
	}
	if f.ClubID != nil {
		sql += fmt.Sprintf(` AND club_id = $%d`, argN)
		args = append(args, int64(*f.ClubID))
		argN++
	}
	if f.City != nil {
		sql += fmt.Sprintf(` AND city = $%d`, argN)
		args = append(args, *f.City)
		argN++
	}

	p := f.Pagination.Normalize()
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, argN, argN+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list players: %w", err)
	}
	defer rows.Close()
	return scanPlayers(rows)
}

func (r *PlayerRepo) Search(ctx context.Context, query string, limit int) ([]*domplayer.Player, error) {
	q := QuerierFromCtx(ctx, r.pool)
	pattern := "%" + query + "%"
	if limit <= 0 {
		limit = 10
	}
	rows, err := q.Query(ctx,
		`SELECT `+playerColumns+` FROM players
		 WHERE status = 'active'
		   AND (full_name ILIKE $1 OR nickname ILIKE $1)
		 ORDER BY full_name
		 LIMIT $2`,
		pattern, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search players: %w", err)
	}
	defer rows.Close()
	return scanPlayers(rows)
}

// ---- Внутренние помощники --------------------------------------------------

func scanPlayer(row pgx.Row) (*domplayer.Player, error) {
	var (
		p           domplayer.Player
		email       string
		nickname    string
		status      string
		stones      float64
		lpTotal     float64
		rating      int
		activeTitle *int64
	)
	err := row.Scan(
		(*int64)(&p.ID),
		&p.FullName,
		&email,
		&p.PasswordHash,
		&nickname,
		&p.KGSNick, &p.OGSNick, &p.FoxNick,
		&p.RFGID, &p.SputnikID,
		&p.City,
		&rating,
		&p.ClubID,
		&stones,
		&lpTotal,
		&status,
		&activeTitle,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan player: %w", err)
	}
	p.Email = shared.Email(email)
	p.Nickname = shared.Nickname(nickname)
	p.Status = domplayer.Status(status)
	p.Rating = shared.Rating(rating)
	p.Stones = shared.Stones(stones)
	p.LPTotal = shared.LP(lpTotal)
	if activeTitle != nil {
		t := shared.TitleID(*activeTitle)
		p.ActiveTitleID = &t
	}
	return &p, nil
}

func scanPlayers(rows pgx.Rows) ([]*domplayer.Player, error) {
	out := make([]*domplayer.Player, 0)
	for rows.Next() {
		p, err := scanPlayer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
