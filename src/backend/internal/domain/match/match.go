package match

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type Status string

const (
	StatusProposed            Status = "proposed"
	StatusAccepted            Status = "accepted"
	StatusPlaying             Status = "playing"
	StatusPendingConfirmation Status = "pending_confirmation"
	StatusConfirmed           Status = "confirmed"
	StatusCancelled           Status = "cancelled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusProposed, StatusAccepted, StatusPlaying,
		StatusPendingConfirmation, StatusConfirmed, StatusCancelled:
		return true
	}
	return false
}

func (s Status) IsFinal() bool {
	return s == StatusConfirmed || s == StatusCancelled
}

type Server string

const (
	ServerOGS     Server = "ogs"
	ServerKGS     Server = "kgs"
	ServerFox     Server = "fox"
	ServerOffline Server = "offline"
)

func (s Server) Valid() bool {
	switch s {
	case ServerOGS, ServerKGS, ServerFox, ServerOffline:
		return true
	}
	return false
}

// Match — партия между двумя игроками в рамках дивизиона сезона.
type Match struct {
	ID         shared.MatchID
	SeasonID   shared.SeasonID
	DivisionID shared.DivisionID

	BlackPlayerID shared.PlayerID
	WhitePlayerID shared.PlayerID
	CreatedBy     shared.PlayerID

	Status Status
	Server Server

	TimeControl string
	Handicap    int
	Komi        float64

	// Результат
	WinnerID     *shared.PlayerID
	ScoreDiff    *float64
	SGFURL       string
	ExternalLink string

	// Аудит
	ResultSubmittedBy *shared.PlayerID
	ConfirmedBy       *shared.PlayerID
	CancelReason      string

	CreatedAt   time.Time
	AcceptedAt  *time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	ConfirmedAt *time.Time
	CancelledAt *time.Time
}

// ---- Конструктор -------------------------------------------------------

type NewMatchParams struct {
	SeasonID      shared.SeasonID
	DivisionID    shared.DivisionID
	BlackPlayerID shared.PlayerID
	WhitePlayerID shared.PlayerID
	CreatedBy     shared.PlayerID
	Server        Server
	TimeControl   string
	Handicap      int
	Komi          float64
}

func New(p NewMatchParams, now time.Time) (*Match, error) {
	if p.BlackPlayerID == p.WhitePlayerID {
		return nil, shared.ErrValidation
	}
	if !p.Server.Valid() {
		return nil, shared.ErrValidation
	}
	if p.CreatedBy != p.BlackPlayerID && p.CreatedBy != p.WhitePlayerID {
		return nil, shared.ErrForbidden
	}
	return &Match{
		SeasonID:      p.SeasonID,
		DivisionID:    p.DivisionID,
		BlackPlayerID: p.BlackPlayerID,
		WhitePlayerID: p.WhitePlayerID,
		CreatedBy:     p.CreatedBy,
		Status:        StatusProposed,
		Server:        p.Server,
		TimeControl:   p.TimeControl,
		Handicap:      p.Handicap,
		Komi:          p.Komi,
		CreatedAt:     now,
	}, nil
}

// ---- Участники / помощники --------------------------------------------

func (m *Match) OpponentOf(p shared.PlayerID) (shared.PlayerID, bool) {
	switch p {
	case m.BlackPlayerID:
		return m.WhitePlayerID, true
	case m.WhitePlayerID:
		return m.BlackPlayerID, true
	}
	return 0, false
}

func (m *Match) IsParticipant(p shared.PlayerID) bool {
	_, ok := m.OpponentOf(p)
	return ok
}

// ---- Переходы состояний ------------------------------------------------

// Accept — соперник принимает вызов.
func (m *Match) Accept(actor shared.PlayerID, now time.Time) error {
	if m.Status != StatusProposed {
		return shared.ErrInvalidState
	}
	opp, ok := m.OpponentOf(actor)
	_ = opp
	if !ok {
		return shared.ErrForbidden
	}
	if actor == m.CreatedBy {
		return shared.ErrForbidden // нельзя принять свой же вызов
	}
	m.Status = StatusAccepted
	m.AcceptedAt = &now
	return nil
}

// Start — любой из участников отмечает начало партии.
func (m *Match) Start(actor shared.PlayerID, now time.Time) error {
	if m.Status != StatusAccepted {
		return shared.ErrInvalidState
	}
	if !m.IsParticipant(actor) {
		return shared.ErrForbidden
	}
	m.Status = StatusPlaying
	m.StartedAt = &now
	return nil
}

// SubmitResult — победитель вносит результат.
// actor должен быть победителем; winner должен совпадать с actor.
func (m *Match) SubmitResult(actor, winner shared.PlayerID, scoreDiff *float64, sgfURL, link string, now time.Time) error {
	if m.Status != StatusPlaying && m.Status != StatusAccepted {
		return shared.ErrInvalidState
	}
	if !m.IsParticipant(actor) {
		return shared.ErrForbidden
	}
	if !m.IsParticipant(winner) {
		return shared.ErrValidation
	}
	if actor != winner {
		return shared.ErrForbidden // результат вносит победитель
	}
	m.WinnerID = &winner
	m.ScoreDiff = scoreDiff
	m.SGFURL = sgfURL
	m.ExternalLink = link
	m.ResultSubmittedBy = &actor
	m.Status = StatusPendingConfirmation
	m.FinishedAt = &now
	return nil
}

// Confirm — соперник подтверждает результат.
func (m *Match) Confirm(actor shared.PlayerID, now time.Time) error {
	if m.Status != StatusPendingConfirmation {
		return shared.ErrInvalidState
	}
	if !m.IsParticipant(actor) {
		return shared.ErrForbidden
	}
	if m.ResultSubmittedBy != nil && actor == *m.ResultSubmittedBy {
		return shared.ErrForbidden // нельзя подтвердить собственный результат
	}
	m.ConfirmedBy = &actor
	m.Status = StatusConfirmed
	m.ConfirmedAt = &now
	return nil
}

// Cancel — игрок или админ отменяет партию.
//
// Игрок: нельзя отменять финальные статусы (confirmed / cancelled),
// и нельзя отменять партию, у которой finished_at старше match_cancel_days.
// Админ: может отменить что угодно, кроме уже отменённой партии.
func (m *Match) Cancel(actor shared.PlayerID, isAdmin bool, cancelAfterDays int, now time.Time) error {
	if isAdmin {
		if m.Status == StatusCancelled {
			return shared.ErrInvalidState
		}
	} else {
		if m.Status.IsFinal() {
			return shared.ErrInvalidState
		}
		if !m.IsParticipant(actor) {
			return shared.ErrForbidden
		}
		if m.FinishedAt != nil {
			deadline := m.FinishedAt.AddDate(0, 0, cancelAfterDays)
			if now.After(deadline) {
				return shared.ErrForbidden
			}
		}
	}
	m.Status = StatusCancelled
	m.CancelledAt = &now
	return nil
}
