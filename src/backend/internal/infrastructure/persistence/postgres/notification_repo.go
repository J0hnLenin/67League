package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/notification"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type NotificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

const notificationColumns = `id, player_id, type, payload, channel, sent_at, read_at, created_at`

func (r *NotificationRepo) Create(ctx context.Context, n *notification.Notification) error {
	q := QuerierFromCtx(ctx, r.pool)
	payloadJSON, err := json.Marshal(n.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	if n.Payload == nil {
		payloadJSON = []byte("{}")
	}
	const sql = `
		INSERT INTO notifications (player_id, type, payload, channel, sent_at, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		n.PlayerID, string(n.Type), payloadJSON, string(n.Channel),
		n.SentAt, n.ReadAt, n.CreatedAt,
	)
	if err := row.Scan((*int64)(&n.ID)); err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *NotificationRepo) GetByID(
	ctx context.Context,
	id shared.NotificationID,
) (*notification.Notification, error) {
	q := QuerierFromCtx(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT `+notificationColumns+` FROM notifications WHERE id = $1`,
		int64(id),
	)
	return scanNotification(row)
}

func (r *NotificationRepo) Update(ctx context.Context, n *notification.Notification) error {
	q := QuerierFromCtx(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE notifications SET sent_at = $2, read_at = $3 WHERE id = $1`,
		int64(n.ID), n.SentAt, n.ReadAt,
	)
	if err != nil {
		return fmt.Errorf("update notification: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *NotificationRepo) ListByPlayer(
	ctx context.Context,
	playerID shared.PlayerID,
	p ports.Pagination,
) ([]*notification.Notification, error) {
	q := QuerierFromCtx(ctx, r.pool)
	pg := p.Normalize()

	rows, err := q.Query(ctx, `
		SELECT `+notificationColumns+`
		FROM notifications
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, int64(playerID), pg.Limit, pg.Offset)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	return scanNotifications(rows)
}

func (r *NotificationRepo) ListUnsent(
	ctx context.Context,
	limit int,
) ([]*notification.Notification, error) {
	q := QuerierFromCtx(ctx, r.pool)
	if limit <= 0 {
		limit = 100
	}
	rows, err := q.Query(ctx, `
		SELECT `+notificationColumns+`
		FROM notifications
		WHERE sent_at IS NULL
		ORDER BY id
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list unsent notifications: %w", err)
	}
	defer rows.Close()
	return scanNotifications(rows)
}

func scanNotification(row pgx.Row) (*notification.Notification, error) {
	var (
		n          notification.Notification
		ntype      string
		channel    string
		payloadRaw []byte
	)
	err := row.Scan(
		(*int64)(&n.ID), &n.PlayerID, &ntype, &payloadRaw,
		&channel, &n.SentAt, &n.ReadAt, &n.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("scan notification: %w", err)
	}
	n.Type = notification.Type(ntype)
	n.Channel = notification.Channel(channel)
	if len(payloadRaw) > 0 {
		_ = json.Unmarshal(payloadRaw, &n.Payload)
	}
	return &n, nil
}

func scanNotifications(rows pgx.Rows) ([]*notification.Notification, error) {
	out := make([]*notification.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
