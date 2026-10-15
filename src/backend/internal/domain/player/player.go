package player

import (
	"errors"
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusBanned   Status = "banned"
)

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusBanned:
		return true
	}
	return false
}

func (s Status) CanPlay() bool       { return s == StatusActive }
func (s Status) CanCreateClub() bool { return s == StatusActive }
func (s Status) CanJoinSeason() bool { return s == StatusActive }

// Player — участник лиги.
type Player struct {
	ID           shared.PlayerID
	FullName     string
	Email        shared.Email
	PasswordHash string
	Nickname     shared.Nickname // ник в лиге

	KGSNick string
	OGSNick string
	FoxNick string

	RFGID     string
	SputnikID string

	City   string
	Rating shared.Rating

	ClubID *shared.ClubID

	Stones  shared.Stones
	LPTotal shared.LP

	Status        Status
	ActiveTitleID *shared.TitleID

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ---- Инварианты / конструктор -----------------------------------------

type NewPlayerParams struct {
	FullName     string
	Email        shared.Email
	PasswordHash string
	Nickname     shared.Nickname
	City         string
	Rating       shared.Rating
	KGSNick      string
	OGSNick      string
	FoxNick      string
	RFGID        string
	SputnikID    string
}

func New(p NewPlayerParams, now time.Time) (*Player, error) {
	if p.FullName == "" || p.City == "" {
		return nil, shared.ErrValidation
	}
	if p.Email == "" || p.Nickname == "" || p.PasswordHash == "" {
		return nil, shared.ErrValidation
	}
	return &Player{
		ID:           shared.NewPlayerID(),
		FullName:     p.FullName,
		Email:        p.Email,
		PasswordHash: p.PasswordHash,
		Nickname:     p.Nickname,
		KGSNick:      p.KGSNick,
		OGSNick:      p.OGSNick,
		FoxNick:      p.FoxNick,
		RFGID:        p.RFGID,
		SputnikID:    p.SputnikID,
		City:         p.City,
		Rating:       p.Rating,
		Stones:       0,
		LPTotal:      0,
		Status:       StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ---- Доменные методы ---------------------------------------------------

func (p *Player) Ban(now time.Time) error {
	if p.Status == StatusBanned {
		return shared.ErrInvalidState
	}
	p.Status = StatusBanned
	p.UpdatedAt = now
	return nil
}

func (p *Player) Unban(now time.Time) error {
	if p.Status != StatusBanned {
		return shared.ErrInvalidState
	}
	p.Status = StatusActive
	p.UpdatedAt = now
	return nil
}

// MarkInactive вызывается по итогам сезона, если игрок сыграл меньше p партий.
func (p *Player) MarkInactive(now time.Time) error {
	if p.Status != StatusActive {
		return shared.ErrInvalidState
	}
	p.Status = StatusInactive
	p.UpdatedAt = now
	return nil
}

// Reactivate — возврат в сезон.
func (p *Player) Reactivate(now time.Time) error {
	if p.Status != StatusInactive {
		return shared.ErrInvalidState
	}
	p.Status = StatusActive
	p.UpdatedAt = now
	return nil
}

func (p *Player) JoinClub(clubID shared.ClubID, now time.Time) error {
	if !p.Status.CanPlay() {
		return shared.ErrForbidden
	}
	if p.ClubID != nil && *p.ClubID == clubID {
		return shared.ErrConflict
	}
	p.ClubID = &clubID
	p.UpdatedAt = now
	return nil
}

func (p *Player) LeaveClub(now time.Time) error {
	if p.ClubID == nil {
		return shared.ErrInvalidState
	}
	p.ClubID = nil
	p.UpdatedAt = now
	return nil
}

func (p *Player) SetTitle(titleID *shared.TitleID, now time.Time) {
	p.ActiveTitleID = titleID
	p.UpdatedAt = now
}

func (p *Player) AddStones(amount shared.Stones, now time.Time) error {
	if amount == 0 {
		return nil
	}
	if p.Stones+amount < 0 {
		return errors.New("insufficient stones")
	}
	p.Stones += amount
	p.UpdatedAt = now
	return nil
}

func (p *Player) AddLP(amount shared.LP, now time.Time) {
	p.LPTotal += amount
	p.UpdatedAt = now
}

func (p *Player) CanCreateMatchWith(opponentID shared.PlayerID) error {
	if !p.Status.CanPlay() {
		return shared.ErrForbidden
	}
	if p.ID == opponentID {
		return shared.ErrValidation
	}
	return nil
}
