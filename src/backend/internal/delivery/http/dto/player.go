package dto

type PlayerResponse struct {
	ID          int64   `json:"id"`
	FullName    string  `json:"full_name"`
	Nickname    string  `json:"nickname"`
	City        string  `json:"city"`
	Rating      int     `json:"rating"`
	ClubID      *int64  `json:"club_id,omitempty"`
	Stones      float64 `json:"stones"`
	LPTotal     float64 `json:"lp_total"`
	Status      string  `json:"status"`
	ActiveTitle *int64  `json:"active_title_id,omitempty"`
}
