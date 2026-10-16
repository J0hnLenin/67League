package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFullFlow — сквозной сценарий: регистрация → вход → клуб → сезон → /me.
func TestFullFlow(t *testing.T) {
	truncateAll(t)

	// 1. Регистрация + вход.
	pid, token := registerAndLogin(t, "alice@example.com", "alice")

	// 2. Создаём клуб.
	createClub := do(t, "POST", "/api/v1/clubs", token, map[string]any{
		"name": "Shibuya",
		"city": "Tokyo",
	}).RequireStatus(t, 201)
	var club struct {
		ID int64 `json:"id"`
	}
	createClub.JSON(t, &club)

	// 3. Сидируем сезон и вступаем.
	seasonID, _, _, divB := seedSeasonWithDivisions(t)
	do(t, "POST", "/api/v1/seasons/"+itoa(seasonID)+"/join", token, nil).
		RequireStatus(t, 200)

	// 4. /me возвращает клуб, а игрок виден в дивизионе.
	me := do(t, "GET", "/api/v1/me", token, nil).RequireStatus(t, 200)
	var profile struct {
		ID       int64  `json:"id"`
		ClubID   int64  `json:"club_id"`
		Nickname string `json:"nickname"`
	}
	me.JSON(t, &profile)
	require.Equal(t, pid, profile.ID)
	require.Equal(t, club.ID, profile.ClubID)
	require.Equal(t, "alice", profile.Nickname)

	members := do(t, "GET", "/api/v1/divisions/"+itoa(divB)+"/members", "", nil).
		RequireStatus(t, 200)
	var list []map[string]any
	members.JSON(t, &list)
	require.Len(t, list, 1)
}
