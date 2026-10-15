package badge

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// PlayerBadge — выданный игроку бейдж. Личные — навсегда.
// Сезонные — с season_id; is_active указывает, показывать ли в текущем сезоне.
type PlayerBadge struct {
	ID        shared.PlayerBadgeID
	PlayerID  shared.PlayerID
	BadgeID   shared.BadgeID
	SeasonID  *shared.SeasonID
	AwardedAt time.Time
	IsActive  bool
}

func NewPlayerBadge(playerID shared.PlayerID, badgeID shared.BadgeID, seasonID *shared.SeasonID, now time.Time) *PlayerBadge {
	return &PlayerBadge{
		PlayerID:  playerID,
		BadgeID:   badgeID,
		SeasonID:  seasonID,
		AwardedAt: now,
		IsActive:  true,
	}
}

// ClubBadge — бейдж, выданный клубу.
type ClubBadge struct {
	ID        shared.ClubBadgeID
	ClubID    shared.ClubID
	BadgeID   shared.BadgeID
	SeasonID  *shared.SeasonID
	AwardedAt time.Time
}

func NewClubBadge(clubID shared.ClubID, badgeID shared.BadgeID, seasonID *shared.SeasonID, now time.Time) *ClubBadge {
	return &ClubBadge{
		ClubID:    clubID,
		BadgeID:   badgeID,
		SeasonID:  seasonID,
		AwardedAt: now,
	}
}

// SuperTitle — история вручения «Чемпиона 67».
type SuperTitle struct {
	ID        shared.SuperTitleID
	SeasonID  shared.SeasonID
	PlayerID  shared.PlayerID
	ClubID    shared.ClubID
	AwardedAt time.Time
}

func NewSuperTitle(seasonID shared.SeasonID, playerID shared.PlayerID, clubID shared.ClubID, now time.Time) *SuperTitle {
	return &SuperTitle{
		SeasonID:  seasonID,
		PlayerID:  playerID,
		ClubID:    clubID,
		AwardedAt: now,
	}
}
