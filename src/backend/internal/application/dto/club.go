package dto

import (
	"github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type CreateClubInput struct {
	Name       string
	City       string
	University string
	CreatorID  shared.PlayerID
}

type CreateClubOutput struct {
	Club *club.Club
}

type JoinClubInput struct {
	PlayerID shared.PlayerID
	ClubID   shared.ClubID
}

type JoinClubOutput struct {
	Player           *player.Player
	MemberCount      int
	FullBadgeAwarded bool
}
