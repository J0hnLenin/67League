package match

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type EventType string

const (
	EventCreated         EventType = "created"
	EventAccepted        EventType = "accepted"
	EventStarted         EventType = "started"
	EventResultSubmitted EventType = "result_submitted"
	EventConfirmed       EventType = "confirmed"
	EventCancelled       EventType = "cancelled"
)

// MatchEvent — запись в журнале изменений партии.
type MatchEvent struct {
	ID        shared.MatchEventID
	MatchID   shared.MatchID
	Type      EventType
	ActorID   shared.PlayerID
	Payload   map[string]any
	CreatedAt time.Time
}

func NewEvent(matchID shared.MatchID, t EventType, actor shared.PlayerID, payload map[string]any, now time.Time) *MatchEvent {
	if payload == nil {
		payload = map[string]any{}
	}
	return &MatchEvent{
		MatchID:   matchID,
		Type:      t,
		ActorID:   actor,
		Payload:   payload,
		CreatedAt: now,
	}
}
