package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeasons_List(t *testing.T) {
	truncateAll(t)

	// Пустая БД.
	resp := do(t, "GET", "/api/v1/seasons", "", nil).RequireStatus(t, 200)
	var empty []any
	resp.JSON(t, &empty)
	require.Empty(t, empty)

	// Сидируем сезон.
	seedSeasonWithDivisions(t)

	list := do(t, "GET", "/api/v1/seasons", "", nil).RequireStatus(t, 200)
	var seasons []map[string]any
	list.JSON(t, &seasons)
	require.Len(t, seasons, 1)
	require.Equal(t, "active", seasons[0]["status"])
}

func TestSeasons_Current(t *testing.T) {
	truncateAll(t)

	// До создания сезона — 404.
	do(t, "GET", "/api/v1/seasons/current", "", nil).RequireStatus(t, 404)

	seedSeasonWithDivisions(t)

	resp := do(t, "GET", "/api/v1/seasons/current", "", nil).RequireStatus(t, 200)
	var s struct {
		Number int    `json:"number"`
		Status string `json:"status"`
	}
	resp.JSON(t, &s)
	require.Equal(t, 1, s.Number)
	require.Equal(t, "active", s.Status)
}

func TestJoinSeason(t *testing.T) {
	truncateAll(t)
	_, token := registerAndLogin(t, "alice@example.com", "alice")
	seasonID, divS, _, divB := seedSeasonWithDivisions(t)
	_ = divS
	_ = divB

	// У alice рейтинг 0 → попадёт в низший дивизион (B).
	join := do(t, "POST", "/api/v1/seasons/"+itoa(seasonID)+"/join", token, nil).
		RequireStatus(t, 200)

	var joined struct {
		DivisionID int64 `json:"division_id"`
		MemberID   int64 `json:"member_id"`
	}
	join.JSON(t, &joined)
	require.NotZero(t, joined.MemberID)
	require.Equal(t, divB, joined.DivisionID)

	// Повторное вступление — 409.
	do(t, "POST", "/api/v1/seasons/"+itoa(seasonID)+"/join", token, nil).
		RequireStatus(t, 409)
}

func TestJoinSeason_NotFound(t *testing.T) {
	truncateAll(t)
	_, token := registerAndLogin(t, "alice@example.com", "alice")

	do(t, "POST", "/api/v1/seasons/99999/join", token, nil).
		RequireStatus(t, 404)
}

func TestDivisionMembers(t *testing.T) {
	truncateAll(t)
	_, tokenA := registerAndLogin(t, "alice@example.com", "alice")
	_, tokenB := registerAndLogin(t, "bob@example.com", "bob")

	seasonID, _, _, divB := seedSeasonWithDivisions(t)

	do(t, "POST", "/api/v1/seasons/"+itoa(seasonID)+"/join", tokenA, nil).RequireStatus(t, 200)
	do(t, "POST", "/api/v1/seasons/"+itoa(seasonID)+"/join", tokenB, nil).RequireStatus(t, 200)

	resp := do(t, "GET", "/api/v1/divisions/"+itoa(divB)+"/members", "", nil).
		RequireStatus(t, 200)

	var members []map[string]any
	resp.JSON(t, &members)
	require.Len(t, members, 2)
}
