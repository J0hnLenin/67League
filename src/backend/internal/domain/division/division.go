package division

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusFinished Status = "finished"
)

type Division struct {
	ID        shared.DivisionID
	SeasonID  shared.SeasonID
	Code      string // "S", "A", "B" или любое другое
	Alias     string // полное название
	RatingMin shared.Rating
	RatingMax shared.Rating
	Order     int // S=0, A=1, B=2 — для сортировки и определения «самого низкого»
	Status    Status

	// Ротация для этого дивизиона в этом сезоне. Настраивается админом.
	RotationUp   int
	RotationDown int

	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(
	seasonID shared.SeasonID,
	code, alias string,
	rmin, rmax shared.Rating,
	order int,
	now time.Time,
) (*Division, error) {
	if code == "" || order < 0 {
		return nil, shared.ErrValidation
	}
	if rmax > 0 && rmax < rmin {
		return nil, shared.ErrValidation
	}
	return &Division{
		ID:        shared.NewDivisionID(),
		SeasonID:  seasonID,
		Code:      code,
		Alias:     alias,
		RatingMin: rmin,
		RatingMax: rmax,
		Order:     order,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// MatchesRating — попадает ли рейтинг игрока в диапазон дивизиона.
func (d *Division) MatchesRating(r shared.Rating) bool {
	if r < d.RatingMin {
		return false
	}
	if d.RatingMax > 0 && r > d.RatingMax {
		return false
	}
	return true
}
