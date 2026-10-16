package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/economy"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type PurchaseRepo struct {
	pool *pgxpool.Pool
}

func NewPurchaseRepo(pool *pgxpool.Pool) *PurchaseRepo {
	return &PurchaseRepo{pool: pool}
}

func (r *PurchaseRepo) Create(ctx context.Context, p *economy.Purchase) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		INSERT INTO purchases (player_id, item_code, price, created_at)
		VALUES ($1, $2, $3, $4) RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		p.PlayerID, p.ItemCode, float64(p.Price), p.CreatedAt,
	)
	if err := row.Scan((*int64)(&p.ID)); err != nil {
		return fmt.Errorf("create purchase: %w", err)
	}
	return nil
}

func (r *PurchaseRepo) ListByPlayer(
	ctx context.Context,
	playerID shared.PlayerID,
	p ports.Pagination,
) ([]*economy.Purchase, error) {
	q := QuerierFromCtx(ctx, r.pool)
	pg := p.Normalize()

	rows, err := q.Query(ctx, `
		SELECT id, player_id, item_code, price, created_at
		FROM purchases
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, int64(playerID), pg.Limit, pg.Offset)
	if err != nil {
		return nil, fmt.Errorf("list purchases: %w", err)
	}
	defer rows.Close()

	out := make([]*economy.Purchase, 0)
	for rows.Next() {
		var (
			p     economy.Purchase
			price float64
		)
		if err := rows.Scan(
			(*int64)(&p.ID), &p.PlayerID, &p.ItemCode, &price, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan purchase: %w", err)
		}
		p.Price = shared.Stones(price)
		out = append(out, &p)
	}
	return out, rows.Err()
}
