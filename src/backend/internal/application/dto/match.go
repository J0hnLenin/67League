package dto

import (
	"github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type ConfirmResultInput struct {
	MatchID shared.MatchID
	ActorID shared.PlayerID
}

type ConfirmResultOutput struct {
	Match        *match.Match
	WinnerLP     shared.LP
	LoserLP      shared.LP
	WinnerStones shared.Stones
	LoserStones  shared.Stones
	StreakBroken bool
}
