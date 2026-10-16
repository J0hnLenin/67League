package dto

type ClubResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	City       string `json:"city"`
	University string `json:"university"`
	LogoURL    string `json:"logo_url,omitempty"`
	CreatedBy  *int64 `json:"created_by,omitempty"`
	Status     string `json:"status"`
	Members    int    `json:"members,omitempty"`
}

type CreateClubRequest struct {
	Name       string `json:"name"`
	City       string `json:"city,omitempty"`
	University string `json:"university,omitempty"`
}

type JoinClubResponse struct {
	ClubID           int64 `json:"club_id"`
	MemberCount      int   `json:"member_count"`
	FullBadgeAwarded bool  `json:"full_badge_awarded"`
}
