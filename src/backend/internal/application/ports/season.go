package ports

import (
	"context"

	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// SeasonRepository — доступ к сезонам.
type SeasonRepository interface {
	Create(ctx context.Context, s *season.Season) error
	GetByID(ctx context.Context, id shared.SeasonID) (*season.Season, error)
	Update(ctx context.Context, s *season.Season) error

	// GetCurrent — сезон, который сейчас не в архиве и не завершён.
	// Если такого нет — ErrNotFound.
	GetCurrent(ctx context.Context) (*season.Season, error)

	// GetLatest — последний по номеру сезон независимо от статуса.
	GetLatest(ctx context.Context) (*season.Season, error)

	List(ctx context.Context, f SeasonFilter) ([]*season.Season, error)
}

type SeasonFilter struct {
	Status *season.Status

	Pagination Pagination
}

// DivisionRepository — доступ к дивизионам сезона.
type DivisionRepository interface {
	Create(ctx context.Context, d *division.Division) error
	GetByID(ctx context.Context, id shared.DivisionID) (*division.Division, error)
	Update(ctx context.Context, d *division.Division) error

	// ListBySeason возвращает дивизионы сезона, отсортированные по Order.
	ListBySeason(ctx context.Context, seasonID shared.SeasonID) ([]*division.Division, error)

	// FindForRating ищет дивизион, в диапазон которого попадает рейтинг.
	FindForRating(ctx context.Context, seasonID shared.SeasonID, r shared.Rating) (*division.Division, error)

	// GetLowest — дивизион с максимальным Order (самый низкий).
	GetLowest(ctx context.Context, seasonID shared.SeasonID) (*division.Division, error)

	// GetHighest — дивизион с минимальным Order (S).
	GetHighest(ctx context.Context, seasonID shared.SeasonID) (*division.Division, error)
}

// DivisionMemberRepository — участия игроков в дивизионах.
type DivisionMemberRepository interface {
	Create(ctx context.Context, m *division.DivisionMember) error
	GetByID(ctx context.Context, id shared.DivisionMemberID) (*division.DivisionMember, error)
	Update(ctx context.Context, m *division.DivisionMember) error

	// GetByPlayerAndSeason — участие игрока в сезоне (через дивизион).
	GetByPlayerAndSeason(ctx context.Context, playerID shared.PlayerID, seasonID shared.SeasonID) (*division.DivisionMember, error)

	// ListByDivision — все участники дивизиона.
	ListByDivision(ctx context.Context, divisionID shared.DivisionID) ([]*division.DivisionMember, error)

	// ListByPlayer — все участия игрока (для истории).
	ListByPlayer(ctx context.Context, playerID shared.PlayerID) ([]*division.DivisionMember, error)
}
