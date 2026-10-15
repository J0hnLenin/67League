package season

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type Status string

const (
	StatusAnnounced    Status = "announced"
	StatusRegistration Status = "registration"
	StatusActive       Status = "active"
	StatusClosing      Status = "closing"
	StatusFinished     Status = "finished"
	StatusArchived     Status = "archived"
)

func (s Status) Valid() bool {
	switch s {
	case StatusAnnounced, StatusRegistration, StatusActive,
		StatusClosing, StatusFinished, StatusArchived:
		return true
	}
	return false
}

// Может ли игрок вступить в сезон сейчас.
func (s Status) CanJoin() bool {
	return s == StatusAnnounced || s == StatusRegistration || s == StatusActive
}

// Можно ли создавать / принимать партии.
func (s Status) CanCreateMatch() bool { return s == StatusActive }

// Можно ли подтверждать уже сыгранные партии.
func (s Status) CanConfirmMatch() bool {
	return s == StatusActive || s == StatusClosing
}

type Season struct {
	ID        shared.SeasonID
	Number    int
	Title     string
	StartDate time.Time
	EndDate   time.Time
	Status    Status
	Settings  Settings
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(number int, title string, start, end time.Time, s Settings, now time.Time) (*Season, error) {
	if number <= 0 || title == "" {
		return nil, shared.ErrValidation
	}
	if !end.After(start) {
		return nil, shared.ErrValidation
	}
	return &Season{
		Number:    number,
		Title:     title,
		StartDate: start,
		EndDate:   end,
		Status:    StatusAnnounced,
		Settings:  s,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ---- Переходы состояний (вызываются админом) --------------------------

func (s *Season) StartRegistration(now time.Time) error {
	if s.Status != StatusAnnounced {
		return shared.ErrInvalidState
	}
	s.Status = StatusRegistration
	s.UpdatedAt = now
	return nil
}

func (s *Season) Start(now time.Time) error {
	if s.Status != StatusRegistration && s.Status != StatusAnnounced {
		return shared.ErrInvalidState
	}
	s.Status = StatusActive
	s.UpdatedAt = now
	return nil
}

func (s *Season) Close(now time.Time) error {
	if s.Status != StatusActive {
		return shared.ErrInvalidState
	}
	s.Status = StatusClosing
	s.UpdatedAt = now
	return nil
}

func (s *Season) Finish(now time.Time) error {
	if s.Status != StatusClosing && s.Status != StatusActive {
		return shared.ErrInvalidState
	}
	s.Status = StatusFinished
	s.UpdatedAt = now
	return nil
}

func (s *Season) Archive(now time.Time) error {
	if s.Status != StatusFinished {
		return shared.ErrInvalidState
	}
	s.Status = StatusArchived
	s.UpdatedAt = now
	return nil
}

// ---- Ограничения, зависящие от сезона ---------------------------------

func (s *Season) MaxMatchesWithOpponent() int { return s.Settings.N }
func (s *Season) MinMatchesForCredit() int    { return s.Settings.P }

// IsDateWithinSeason проверяет, что действие укладывается в окно сезона.
func (s *Season) IsDateWithinSeason(t time.Time) bool {
	return !t.Before(s.StartDate) && !t.After(s.EndDate)
}
