package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type BadgeRepo struct {
	pool *pgxpool.Pool
}

func NewBadgeRepo(pool *pgxpool.Pool) *BadgeRepo { return &BadgeRepo{pool: pool} }

const badgeColumns = `id, code, icon, name, description, type, condition, rarity, title_id, created_at`

func (r *BadgeRepo) Create(ctx context.Context, b *dombadge.Badge) error {
	q := QuerierFromCtx(ctx, r.pool)
	condJSON, err := json.Marshal(b.Condition)
	if err != nil {
		return fmt.Errorf("marshal condition: %w", err)
	}
	if b.Condition == nil {
		condJSON = []byte("{}")
	}
	const sql = `
		INSERT INTO badges (code, icon, name, description, type, condition, rarity, title_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		b.Code, b.Icon, b.Name, b.Description,
		string(b.Type), condJSON, string(b.Rarity), b.TitleID, b.CreatedAt,
	)
	if err := row.Scan((*int64)(&b.ID)); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create badge: %w", shared.ErrAlreadyExists)
		}
		return fmt.Errorf("create badge: %w", err)
	}
	return nil
}

func (r *BadgeRepo) GetByID(ctx context.Context, id shared.BadgeID) (*dombadge.Badge, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+badgeColumns+` FROM badges WHERE id = $1`, int64(id))
	return scanBadge(row)
}

func (r *BadgeRepo) GetByCode(ctx context.Context, code string) (*dombadge.Badge, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+badgeColumns+` FROM badges WHERE code = $1`, code)
	return scanBadge(row)
}

func (r *BadgeRepo) List(ctx context.Context) ([]*dombadge.Badge, error) {
	q := QuerierFromCtx(ctx, r.pool)
	rows, err := q.Query(ctx, `SELECT `+badgeColumns+` FROM badges ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list badges: %w", err)
	}
	defer rows.Close()

	out := make([]*dombadge.Badge, 0)
	for rows.Next() {
		b, err := scanBadge(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func scanBadge(row pgx.Row) (*dombadge.Badge, error) {
	var (
		b            dombadge.Badge
		badgeType    string
		rarity       string
		conditionRaw []byte
	)
	err := row.Scan(
		(*int64)(&b.ID), &b.Code, &b.Icon, &b.Name, &b.Description,
		&badgeType, &conditionRaw, &rarity, &b.TitleID, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan badge: %w", err)
	}
	b.Type = dombadge.Type(badgeType)
	b.Rarity = dombadge.Rarity(rarity)
	if len(conditionRaw) > 0 {
		_ = json.Unmarshal(conditionRaw, &b.Condition)
	}
	return &b, nil
}
