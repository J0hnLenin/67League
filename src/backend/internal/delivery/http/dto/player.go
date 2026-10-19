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

type PlayerBadgeResponse struct {
	ID        int64  `json:"id"`
	BadgeID   int64  `json:"badge_id"`
	Code      string `json:"code"`
	Icon      string `json:"icon"`
	Name      string `json:"name"`
	SeasonID  *int64 `json:"season_id,omitempty"`
	AwardedAt string `json:"awarded_at"`
	IsActive  bool   `json:"is_active"`
}

type TitleResponse struct {
	ID      int64  `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	BadgeID *int64 `json:"badge_id,omitempty"`
}

type StoneTransactionResponse struct {
	ID        int64   `json:"id"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
	CreatedAt string  `json:"created_at"`
}

type PlayerSearchResponse struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Nickname string `json:"nickname"`
}
