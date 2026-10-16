package dto

type SeasonResponse struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    string `json:"status"`
}

type DivisionResponse struct {
	ID        int64  `json:"id"`
	SeasonID  int64  `json:"season_id"`
	Code      string `json:"code"`
	Alias     string `json:"alias,omitempty"`
	RatingMin int    `json:"rating_min"`
	RatingMax int    `json:"rating_max"`
	Order     int    `json:"order"`
	Status    string `json:"status"`
}

type DivisionMemberResponse struct {
	PlayerID      int64   `json:"player_id"`
	DivisionID    int64   `json:"division_id"`
	LP            float64 `json:"lp"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	MatchesPlayed int     `json:"matches_played"`
	StreakCurrent int     `json:"streak_current"`
	StreakMax     int     `json:"streak_max"`
	FinalPlace    *int    `json:"final_place,omitempty"`
	Status        string  `json:"status"`
}

type JoinSeasonResponse struct {
	DivisionID int64 `json:"division_id"`
	MemberID   int64 `json:"member_id"`
}
