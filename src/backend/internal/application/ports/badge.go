package ports

import (
	"context"

	"github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// BadgeRepository — справочник бейджей.
type BadgeRepository interface {
	Create(ctx context.Context, b *badge.Badge) error
	GetByID(ctx context.Context, id shared.BadgeID) (*badge.Badge, error)
	GetByCode(ctx context.Context, code string) (*badge.Badge, error)
	List(ctx context.Context) ([]*badge.Badge, error)
}

// TitleRepository — справочник титулов.
type TitleRepository interface {
	Create(ctx context.Context, t *badge.Title) error
	GetByID(ctx context.Context, id shared.TitleID) (*badge.Title, error)
	GetByCode(ctx context.Context, code string) (*badge.Title, error)
	List(ctx context.Context) ([]*badge.Title, error)
}

// PlayerBadgeRepository — выданные игрокам бейджи.
type PlayerBadgeRepository interface {
	Create(ctx context.Context, pb *badge.PlayerBadge) error

	ListByPlayer(ctx context.Context, playerID shared.PlayerID) ([]*badge.PlayerBadge, error)
	ListByPlayerAndSeason(ctx context.Context, playerID shared.PlayerID, seasonID shared.SeasonID) ([]*badge.PlayerBadge, error)

	// Has — выдан ли уже бейдж (для идемпотентности автовыдачи).
	Has(ctx context.Context, playerID shared.PlayerID, badgeID shared.BadgeID) (bool, error)

	// DeactivateSeasonal — при старте нового сезона снимает is_active у прошлых сезонных.
	DeactivateSeasonal(ctx context.Context, playerID shared.PlayerID) error
}

// ClubBadgeRepository — выданные клубам бейджи.
type ClubBadgeRepository interface {
	Create(ctx context.Context, cb *badge.ClubBadge) error
	ListByClub(ctx context.Context, clubID shared.ClubID) ([]*badge.ClubBadge, error)
	Has(ctx context.Context, clubID shared.ClubID, badgeID shared.BadgeID) (bool, error)
}

// SuperTitleRepository — история «Чемпиона 67».
type SuperTitleRepository interface {
	Create(ctx context.Context, st *badge.SuperTitle) error
	GetBySeason(ctx context.Context, seasonID shared.SeasonID) (*badge.SuperTitle, error)
	List(ctx context.Context, f SuperTitleFilter) ([]*badge.SuperTitle, error)
}

type SuperTitleFilter struct {
	PlayerID *shared.PlayerID
	ClubID   *shared.ClubID

	Pagination Pagination
}
