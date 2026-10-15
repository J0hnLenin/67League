package ports

import (
	"context"

	"github.com/J0hnLenin/67League/internal/domain/admin"
	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// PlayerRepository — доступ к игрокам.
type PlayerRepository interface {
	Create(ctx context.Context, p *player.Player) error

	GetByID(ctx context.Context, id shared.PlayerID) (*player.Player, error)
	GetByEmail(ctx context.Context, email shared.Email) (*player.Player, error)
	GetByNickname(ctx context.Context, nick shared.Nickname) (*player.Player, error)

	Update(ctx context.Context, p *player.Player) error

	ExistsByEmail(ctx context.Context, email shared.Email) (bool, error)
	ExistsByNickname(ctx context.Context, nick shared.Nickname) (bool, error)

	// ListByClub — все игроки клуба.
	ListByClub(ctx context.Context, clubID shared.ClubID) ([]*player.Player, error)

	// CountByClub — число участников клуба (для бейджа «Полный состав»).
	CountByClub(ctx context.Context, clubID shared.ClubID) (int, error)

	// List — общий список с фильтром (для админки).
	List(ctx context.Context, f PlayerFilter) ([]*player.Player, error)
}

// PlayerFilter — фильтр для админского списка.
type PlayerFilter struct {
	Status *player.Status
	ClubID *shared.ClubID
	City   *string

	Pagination Pagination
}

// AdminRepository — доступ к админам. Отдельная сущность, отдельный вход.
type AdminRepository interface {
	Create(ctx context.Context, a *admin.Admin) error
	GetByID(ctx context.Context, id shared.AdminID) (*admin.Admin, error)
	GetByLogin(ctx context.Context, login string) (*admin.Admin, error)
	Update(ctx context.Context, a *admin.Admin) error
}
