package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/economy"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type StoneTransactionRepo struct {
	pool *pgxpool.Pool
}

func NewStoneTransactionRepo(pool *pgxpool.Pool) *StoneTransactionRepo {
	return &StoneTransactionRepo{pool: pool}
}

func (r *StoneTransactionRepo) Create(ctx context.Context, t *economy.StoneTransaction) error {
	q := QuerierFromCtx(ctx, r.pool)
	const sql = `
		INSERT INTO stone_transactions (player_id, amount, reason, ref_id, created_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		t.PlayerID, float64(t.Amount), t.Reason, t.RefID, t.CreatedAt,
	)
	if err := row.Scan((*int64)(&t.ID)); err != nil {
		return fmt.Errorf("create stone tx: %w", err)
	}
	return nil
}

func (r *StoneTransactionRepo) ListByPlayer(
	ctx context.Context,
	playerID shared.PlayerID,
	p ports.Pagination,
) ([]*economy.StoneTransaction, error) {
	q := QuerierFromCtx(ctx, r.pool)
	pg := p.Normalize()

	rows, err := q.Query(ctx, `
		SELECT id, player_id, amount, reason, ref_id, created_at
		FROM stone_transactions
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, int64(playerID), pg.Limit, pg.Offset)
	if err != nil {
		return nil, fmt.Errorf("list stone tx: %w", err)
	}
	defer rows.Close()

	out := make([]*economy.StoneTransaction, 0)
	for rows.Next() {
		var (
			t      economy.StoneTransaction
			amount float64
		)
		if err := rows.Scan(
			(*int64)(&t.ID), &t.PlayerID, &amount, &t.Reason, &t.RefID, &t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan stone tx: %w", err)
		}
		t.Amount = shared.Stones(amount)
		out = append(out, &t)
	}
	return out, rows.Err()
}
