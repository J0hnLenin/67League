package notification

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type Type string

const (
	TypeNewChallenge      Type = "new_challenge"
	TypeChallengeAccepted Type = "challenge_accepted"
	TypeResultSubmitted   Type = "result_submitted"
	TypeResultConfirmed   Type = "result_confirmed"
	TypeBadgeAwarded      Type = "badge_awarded"
	TypeSeasonStarted     Type = "season_started"
	TypeSeasonFinished    Type = "season_finished"
)

type Channel string

const (
	ChannelEmail    Channel = "email"
	ChannelTelegram Channel = "telegram"
	ChannelVK       Channel = "vk"
)

// Notification — уведомление игроку.
type Notification struct {
	ID        shared.NotificationID
	PlayerID  shared.PlayerID
	Type      Type
	Payload   map[string]any
	Channel   Channel
	SentAt    *time.Time
	ReadAt    *time.Time
	CreatedAt time.Time
}

func New(playerID shared.PlayerID, t Type, payload map[string]any, ch Channel, now time.Time) *Notification {
	if payload == nil {
		payload = map[string]any{}
	}
	return &Notification{
		PlayerID:  playerID,
		Type:      t,
		Payload:   payload,
		Channel:   ch,
		CreatedAt: now,
	}
}

func (n *Notification) MarkSent(now time.Time) { n.SentAt = &now }
func (n *Notification) MarkRead(now time.Time) { n.ReadAt = &now }
