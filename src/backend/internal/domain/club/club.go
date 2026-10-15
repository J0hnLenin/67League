package club

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

type Club struct {
	ID         shared.ClubID
	Name       string
	City       string
	University string
	LogoURL    string
	CreatedBy  shared.PlayerID
	Status     Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func New(name, city, university string, createdBy shared.PlayerID, now time.Time) (*Club, error) {
	if name == "" {
		return nil, shared.ErrValidation
	}
	return &Club{
		ID:         shared.NewClubID(),
		Name:       name,
		City:       city,
		University: university,
		CreatedBy:  createdBy,
		Status:     StatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (c *Club) Archive(now time.Time) error {
	if c.Status == StatusArchived {
		return shared.ErrInvalidState
	}
	c.Status = StatusArchived
	c.UpdatedAt = now
	return nil
}
