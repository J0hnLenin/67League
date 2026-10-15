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

type CreateMatchInput struct {
	SeasonID      shared.SeasonID
	DivisionID    shared.DivisionID
	BlackPlayerID shared.PlayerID
	WhitePlayerID shared.PlayerID
	CreatedBy     shared.PlayerID
	Server        match.Server
	TimeControl   string
	Handicap      int
	Komi          float64
}

type CreateMatchOutput struct {
	Match *match.Match
}

type AcceptMatchInput struct {
	MatchID shared.MatchID
	ActorID shared.PlayerID
}

type AcceptMatchOutput struct {
	Match *match.Match
}

type StartMatchInput struct {
	MatchID shared.MatchID
	ActorID shared.PlayerID
}

type StartMatchOutput struct {
	Match *match.Match
}

type SubmitResultInput struct {
	MatchID   shared.MatchID
	ActorID   shared.PlayerID
	WinnerID  shared.PlayerID
	ScoreDiff *float64
	SGFURL    string
	Link      string
}

type SubmitResultOutput struct {
	Match *match.Match
}

type CancelMatchInput struct {
	MatchID shared.MatchID
	ActorID shared.PlayerID
	IsAdmin bool
	Reason  string
}

type CancelMatchOutput struct {
	Match *match.Match
}
