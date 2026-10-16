package dto

type RegisterRequest struct {
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Nickname  string `json:"nickname"`
	City      string `json:"city"`
	KGSNick   string `json:"kgs_nick,omitempty"`
	OGSNick   string `json:"ogs_nick,omitempty"`
	FoxNick   string `json:"fox_nick,omitempty"`
	RFGID     string `json:"rfg_id,omitempty"`
	SputnikID string `json:"sputnik_id,omitempty"`
	ClubID    *int64 `json:"club_id,omitempty"`
}

type RegisterResponse struct {
	PlayerID   int64  `json:"player_id"`
	DivisionID *int64 `json:"division_id,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	PlayerID int64  `json:"player_id"`
	Nickname string `json:"nickname"`
	Token    string `json:"token"`
}
