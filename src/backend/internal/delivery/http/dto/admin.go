package dto

// ---- Auth --------------------------------------------------------------

type AdminLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	AdminID int64  `json:"admin_id"`
	Token   string `json:"token"`
}

// ---- Season ------------------------------------------------------------

type CreateSeasonRequest struct {
	Number    int                    `json:"number"`
	Title     string                 `json:"title"`
	StartDate string                 `json:"start_date"` // RFC3339
	EndDate   string                 `json:"end_date"`
	Settings  *SeasonSettingsRequest `json:"settings,omitempty"`
}

type SeasonSettingsRequest struct {
	K1 int `json:"k1,omitempty"`
	K2 int `json:"k2,omitempty"`
	K3 int `json:"k3,omitempty"`
	P  int `json:"p,omitempty"`
	N  int `json:"n,omitempty"`
	T  int `json:"t,omitempty"`
}

type PatchSeasonRequest struct {
	Title     *string `json:"title,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
}

type FinishSeasonResponse struct {
	Season     SeasonResponse         `json:"season"`
	Standings  []DivisionStandingJSON `json:"standings"`
	ClubTable  []ClubStandingJSON     `json:"club_table"`
	BestClubID *int64                 `json:"best_club_id,omitempty"`
}

type DivisionStandingJSON struct {
	DivisionID int64                `json:"division_id"`
	Code       string               `json:"code"`
	Rows       []MemberStandingJSON `json:"rows"`
}

type MemberStandingJSON struct {
	PlayerID int64   `json:"player_id"`
	LP       float64 `json:"lp"`
	Place    int     `json:"place,omitempty"`
	Excluded bool    `json:"excluded"`
}

type ClubStandingJSON struct {
	ClubID  int64   `json:"club_id"`
	TotalLP float64 `json:"total_lp"`
	Members int     `json:"members"`
	Top5Sum float64 `json:"top5_sum"`
}

type RotationResponse struct {
	Created   int                  `json:"created"`
	Transfers []PlayerTransferJSON `json:"transfers"`
}

type PlayerTransferJSON struct {
	PlayerID  int64  `json:"player_id"`
	FromDivID int64  `json:"from_division_id"`
	ToDivID   int64  `json:"to_division_id"`
	Direction string `json:"direction"`
	FromPlace int    `json:"from_place,omitempty"`
}

type AwardBadgesResponse struct {
	Awarded []PlayerBadgeResponse `json:"awarded"`
}

type AwardSuperTitleResponse struct {
	Awarded  bool   `json:"awarded"`
	PlayerID *int64 `json:"player_id,omitempty"`
	ClubID   *int64 `json:"club_id,omitempty"`
}

// ---- Player ------------------------------------------------------------

type BanPlayerRequest struct {
	Reason string `json:"reason,omitempty"`
}

type GrantBadgeRequest struct {
	BadgeCode string `json:"badge_code"`
}

// ---- Settings ----------------------------------------------------------

type PatchSettingsRequest struct {
	Values map[string]any `json:"values"`
}
