package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type MatchEventRepo struct {
	pool *pgxpool.Pool
}

func NewMatchEventRepo(pool *pgxpool.Pool) *MatchEventRepo {
	return &MatchEventRepo{pool: pool}
}

func (r *MatchEventRepo) Create(ctx context.Context, e *dommatch.MatchEvent) error {
	q := QuerierFromCtx(ctx, r.pool)
	payload := e.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	const sql = `
		INSERT INTO match_events (match_id, event_type, actor_id, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	row := q.QueryRow(ctx, sql,
		e.MatchID, string(e.Type), e.ActorID, payloadJSON, e.CreatedAt,
	)
	if err := row.Scan((*int64)(&e.ID)); err != nil {
		return fmt.Errorf("create match event: %w", err)
	}
	return nil
}

func (r *MatchEventRepo) ListByMatch(ctx context.Context, matchID shared.MatchID) ([]*dommatch.MatchEvent, error) {
	q := QuerierFromCtx(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT id, match_id, event_type, actor_id, payload, created_at
		FROM match_events
		WHERE match_id = $1
		ORDER BY id
	`, int64(matchID))
	if err != nil {
		return nil, fmt.Errorf("list match events: %w", err)
	}
	defer rows.Close()

	out := make([]*dommatch.MatchEvent, 0)
	for rows.Next() {
		var (
			e          dommatch.MatchEvent
			eventType  string
			payloadRaw []byte
		)
		if err := rows.Scan(
			(*int64)(&e.ID), &e.MatchID, &eventType, &e.ActorID,
			&payloadRaw, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		e.Type = dommatch.EventType(eventType)
		if len(payloadRaw) > 0 {
			if err := json.Unmarshal(payloadRaw, &e.Payload); err != nil {
				return nil, fmt.Errorf("unmarshal payload: %w", err)
			}
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}
