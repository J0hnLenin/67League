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

type MatchResponse struct {
	ID            int64    `json:"id"`
	SeasonID      int64    `json:"season_id"`
	DivisionID    int64    `json:"division_id"`
	BlackPlayerID int64    `json:"black_player_id"`
	WhitePlayerID int64    `json:"white_player_id"`
	Status        string   `json:"status"`
	Server        string   `json:"server"`
	WinnerID      *int64   `json:"winner_id,omitempty"`
	ScoreDiff     *float64 `json:"score_diff,omitempty"`
}
