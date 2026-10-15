package dto

import (
	"github.com/J0hnLenin/67League/internal/domain/club"
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

type CalculateSeasonResultsInput struct {
	SeasonID shared.SeasonID
}

type CalculateSeasonResultsOutput struct {
	Season    *season.Season
	Standings []DivisionStanding
	ClubTable []ClubStanding
	BestClub  *ClubStanding
}

type DivisionStanding struct {
	Division *division.Division
	Rows     []MemberStanding
}

type MemberStanding struct {
	Member   *division.DivisionMember
	PlayerID shared.PlayerID
	Place    int
	Excluded bool
}

type ClubStanding struct {
	ClubID  shared.ClubID
	Club    *club.Club
	TotalLP shared.LP
	Members int
	Top5Sum shared.LP
}

type ApplyRotationInput struct {
	PreviousSeasonID shared.SeasonID
	NextSeasonID     shared.SeasonID
}

type ApplyRotationOutput struct {
	Transfers []PlayerTransfer
	Created   int
}

type PlayerTransfer struct {
	PlayerID  shared.PlayerID
	FromDivID shared.DivisionID
	ToDivID   shared.DivisionID
	Direction RotationDirection
	FromPlace int
}

type RotationDirection string

const (
	RotationUp       RotationDirection = "up"
	RotationStay     RotationDirection = "stay"
	RotationDown     RotationDirection = "down"
	RotationByRating RotationDirection = "by_rating"
)
