package ports

import (
	"context"

	"github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// ClubRepository — доступ к клубам.
type ClubRepository interface {
	Create(ctx context.Context, c *club.Club) error

	GetByID(ctx context.Context, id shared.ClubID) (*club.Club, error)
	GetByName(ctx context.Context, name string) (*club.Club, error)

	Update(ctx context.Context, c *club.Club) error

	ExistsByName(ctx context.Context, name string) (bool, error)

	List(ctx context.Context, f ClubFilter) ([]*club.Club, error)
}

// ClubFilter — фильтр списка клубов.
type ClubFilter struct {
	City   *string
	Status *club.Status

	Pagination Pagination
}
