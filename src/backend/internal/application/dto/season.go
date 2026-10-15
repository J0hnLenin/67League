package dto

import (
	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type JoinSeasonInput struct {
	PlayerID shared.PlayerID
	SeasonID shared.SeasonID
}

type JoinSeasonOutput struct {
	Member   *division.DivisionMember
	Division *division.Division
	Season   *season.Season
}
