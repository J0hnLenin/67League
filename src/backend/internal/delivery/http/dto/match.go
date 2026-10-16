package dto

type CreateMatchRequest struct {
	SeasonID      int64   `json:"season_id"`
	DivisionID    int64   `json:"division_id"`
	BlackPlayerID int64   `json:"black_player_id"`
	WhitePlayerID int64   `json:"white_player_id"`
	Server        string  `json:"server"`
	TimeControl   string  `json:"time_control,omitempty"`
	Handicap      int     `json:"handicap,omitempty"`
	Komi          float64 `json:"komi,omitempty"`
}

type SubmitResultRequest struct {
	WinnerID  int64    `json:"winner_id"`
	ScoreDiff *float64 `json:"score_diff,omitempty"`
	SGFURL    string   `json:"sgf_url,omitempty"`
	Link      string   `json:"link,omitempty"`
}

type CancelMatchRequest struct {
	Reason string `json:"reason,omitempty"`
}

type MatchResponse struct {
	ID            int64    `json:"id"`
	SeasonID      int64    `json:"season_id"`
	DivisionID    int64    `json:"division_id"`
	BlackPlayerID int64    `json:"black_player_id"`
	WhitePlayerID int64    `json:"white_player_id"`
	CreatedBy     int64    `json:"created_by"`
	Status        string   `json:"status"`
	Server        string   `json:"server"`
	TimeControl   string   `json:"time_control,omitempty"`
	Handicap      int      `json:"handicap,omitempty"`
	Komi          float64  `json:"komi,omitempty"`
	WinnerID      *int64   `json:"winner_id,omitempty"`
	ScoreDiff     *float64 `json:"score_diff,omitempty"`
	SGFURL        string   `json:"sgf_url,omitempty"`
	Link          string   `json:"link,omitempty"`
	CancelReason  string   `json:"cancel_reason,omitempty"`
	CreatedAt     string   `json:"created_at"`
	AcceptedAt    *string  `json:"accepted_at,omitempty"`
	StartedAt     *string  `json:"started_at,omitempty"`
	FinishedAt    *string  `json:"finished_at,omitempty"`
	ConfirmedAt   *string  `json:"confirmed_at,omitempty"`
	CancelledAt   *string  `json:"cancelled_at,omitempty"`
}
