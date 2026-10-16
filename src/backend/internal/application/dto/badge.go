package dto

import (
	"github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type AwardPersonalBadgesInput struct {
	PlayerID shared.PlayerID
	SeasonID shared.SeasonID
}

type AwardPersonalBadgesOutput struct {
	Awarded []*badge.PlayerBadge
}

type AwardSeasonBadgesInput struct {
	SeasonID shared.SeasonID
}

type AwardSeasonBadgesOutput struct {
	Awarded []*badge.PlayerBadge
}

type AwardSuperTitleInput struct {
	SeasonID shared.SeasonID
}

type AwardSuperTitleOutput struct {
	SuperTitle *badge.SuperTitle
	Player     *shared.PlayerID
	Club       *shared.ClubID
	Awarded    bool
}
