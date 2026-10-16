package dto

import (
	"time"

	domclub "github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/division"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
)

// ---- Match ----------------------------------------------------------------

func MatchToResponse(m *dommatch.Match) MatchResponse {
	resp := MatchResponse{
		ID:            int64(m.ID),
		SeasonID:      int64(m.SeasonID),
		DivisionID:    int64(m.DivisionID),
		BlackPlayerID: int64(m.BlackPlayerID),
		WhitePlayerID: int64(m.WhitePlayerID),
		CreatedBy:     int64(m.CreatedBy),
		Status:        string(m.Status),
		Server:        string(m.Server),
		TimeControl:   m.TimeControl,
		Handicap:      m.Handicap,
		Komi:          m.Komi,
		SGFURL:        m.SGFURL,
		Link:          m.ExternalLink,
		CancelReason:  m.CancelReason,
		CreatedAt:     m.CreatedAt.UTC().Format(time.RFC3339),
	}
	if m.WinnerID != nil {
		id := int64(*m.WinnerID)
		resp.WinnerID = &id
	}
	if m.ScoreDiff != nil {
		resp.ScoreDiff = m.ScoreDiff
	}
	resp.AcceptedAt = fmtTimePtr(m.AcceptedAt)
	resp.StartedAt = fmtTimePtr(m.StartedAt)
	resp.FinishedAt = fmtTimePtr(m.FinishedAt)
	resp.ConfirmedAt = fmtTimePtr(m.ConfirmedAt)
	resp.CancelledAt = fmtTimePtr(m.CancelledAt)
	return resp
}

func MatchesToResponse(ms []*dommatch.Match) []MatchResponse {
	out := make([]MatchResponse, 0, len(ms))
	for _, m := range ms {
		out = append(out, MatchToResponse(m))
	}
	return out
}

// ---- Player ---------------------------------------------------------------

func PlayerToResponse(p *domplayer.Player) PlayerResponse {
	resp := PlayerResponse{
		ID:       int64(p.ID),
		FullName: p.FullName,
		Nickname: string(p.Nickname),
		City:     p.City,
		Rating:   int(p.Rating),
		Stones:   float64(p.Stones),
		LPTotal:  float64(p.LPTotal),
		Status:   string(p.Status),
	}
	if p.ClubID != nil {
		id := int64(*p.ClubID)
		resp.ClubID = &id
	}
	if p.ActiveTitleID != nil {
		id := int64(*p.ActiveTitleID)
		resp.ActiveTitle = &id
	}
	return resp
}

func PlayersToResponse(ps []*domplayer.Player) []PlayerResponse {
	out := make([]PlayerResponse, 0, len(ps))
	for _, p := range ps {
		out = append(out, PlayerToResponse(p))
	}
	return out
}

// ---- Season ---------------------------------------------------------------

func SeasonToResponse(s *domseason.Season) SeasonResponse {
	return SeasonResponse{
		ID:        int64(s.ID),
		Number:    s.Number,
		Title:     s.Title,
		StartDate: s.StartDate.UTC().Format(time.RFC3339),
		EndDate:   s.EndDate.UTC().Format(time.RFC3339),
		Status:    string(s.Status),
	}
}

// ---- Club -----------------------------------------------------------------

func ClubToResponse(c *domclub.Club, members int) ClubResponse {
	resp := ClubResponse{
		ID:         int64(c.ID),
		Name:       c.Name,
		City:       c.City,
		University: c.University,
		LogoURL:    c.LogoURL,
		Status:     string(c.Status),
		Members:    members,
	}
	if c.CreatedBy != 0 {
		id := int64(c.CreatedBy)
		resp.CreatedBy = &id
	}
	return resp
}

// ---- Division -------------------------------------------------------------

func DivisionToResponse(d *division.Division) DivisionResponse {
	return DivisionResponse{
		ID:        int64(d.ID),
		SeasonID:  int64(d.SeasonID),
		Code:      d.Code,
		Alias:     d.Alias,
		RatingMin: int(d.RatingMin),
		RatingMax: int(d.RatingMax),
		Order:     d.Order,
		Status:    string(d.Status),
	}
}

func DivisionMemberToResponse(m *division.DivisionMember) DivisionMemberResponse {
	return DivisionMemberResponse{
		PlayerID:      int64(m.PlayerID),
		DivisionID:    int64(m.DivisionID),
		LP:            float64(m.LP),
		Wins:          m.Wins,
		Losses:        m.Losses,
		MatchesPlayed: m.MatchesPlayed,
		StreakCurrent: m.StreakCurrent,
		StreakMax:     m.StreakMax,
		FinalPlace:    m.FinalPlace,
		Status:        string(m.Status),
	}
}

func DivisionMembersToResponse(ms []*division.DivisionMember) []DivisionMemberResponse {
	out := make([]DivisionMemberResponse, 0, len(ms))
	for _, m := range ms {
		out = append(out, DivisionMemberToResponse(m))
	}
	return out
}

// ---- helpers --------------------------------------------------------------

func fmtTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
