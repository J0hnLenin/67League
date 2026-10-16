package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// setupMatchScenario готовит двух игроков в одном дивизионе активного сезона.
// Возвращает токены, IDs, seasonID и divisionID.
func setupMatchScenario(t *testing.T) (tokA, tokB string, pidA, pidB, seasonID, divID int64) {
	t.Helper()

	pidA, tokA = registerAndLogin(t, "alice@example.com", "alice")
	pidB, tokB = registerAndLogin(t, "bob@example.com", "bob")

	seasonID, _, _, divB := seedSeasonWithDivisions(t)
	divID = divB

	do(t, "POST", "/api/v1/seasons/"+itoa(seasonID)+"/join", tokA, nil).RequireStatus(t, 200)
	do(t, "POST", "/api/v1/seasons/"+itoa(seasonID)+"/join", tokB, nil).RequireStatus(t, 200)

	return tokA, tokB, pidA, pidB, seasonID, divID
}

func TestMatch_FullFlow(t *testing.T) {
	truncateAll(t)
	tokA, tokB, pidA, pidB, seasonID, divID := setupMatchScenario(t)

	// A создаёт вызов на B.
	create := do(t, "POST", "/api/v1/matches", tokA, map[string]any{
		"season_id":       seasonID,
		"division_id":     divID,
		"black_player_id": pidA,
		"white_player_id": pidB,
		"server":          "ogs",
	}).RequireStatus(t, 201)

	var m struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	create.JSON(t, &m)
	require.NotZero(t, m.ID)
	require.Equal(t, "proposed", m.Status)

	// B принимает.
	accept := do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/accept", tokB, nil).
		RequireStatus(t, 200)
	var afterAccept struct {
		Status string `json:"status"`
	}
	accept.JSON(t, &afterAccept)
	require.Equal(t, "accepted", afterAccept.Status)

	// A стартует.
	start := do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/start", tokA, nil).
		RequireStatus(t, 200)
	var afterStart struct {
		Status string `json:"status"`
	}
	start.JSON(t, &afterStart)
	require.Equal(t, "playing", afterStart.Status)

	// A вносит результат: победил A.
	result := do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/result", tokA, map[string]any{
		"winner_id":  pidA,
		"score_diff": 5.5,
	}).RequireStatus(t, 200)
	var afterResult struct {
		Status string `json:"status"`
		Winner int64  `json:"winner_id"`
	}
	result.JSON(t, &afterResult)
	require.Equal(t, "pending_confirmation", afterResult.Status)
	require.Equal(t, pidA, afterResult.Winner)

	// B подтверждает.
	confirm := do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/confirm", tokB, nil).
		RequireStatus(t, 200)
	var afterConfirm struct {
		Status string `json:"status"`
	}
	confirm.JSON(t, &afterConfirm)
	require.Equal(t, "confirmed", afterConfirm.Status)
}

func TestMatch_CannotSubmitResultForOpponent(t *testing.T) {
	truncateAll(t)
	tokA, tokB, pidA, pidB, seasonID, divID := setupMatchScenario(t)

	create := do(t, "POST", "/api/v1/matches", tokA, map[string]any{
		"season_id":       seasonID,
		"division_id":     divID,
		"black_player_id": pidA,
		"white_player_id": pidB,
		"server":          "ogs",
	}).RequireStatus(t, 201)

	var m struct {
		ID int64 `json:"id"`
	}
	create.JSON(t, &m)

	do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/accept", tokB, nil).RequireStatus(t, 200)
	do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/start", tokA, nil).RequireStatus(t, 200)

	// A пытается внести результат, где победитель — B. Use case запрещает.
	do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/result", tokA, map[string]any{
		"winner_id": pidB,
	}).RequireStatus(t, 403)
}

func TestMatch_CannotConfirmOwnResult(t *testing.T) {
	truncateAll(t)
	tokA, tokB, pidA, pidB, seasonID, divID := setupMatchScenario(t)

	create := do(t, "POST", "/api/v1/matches", tokA, map[string]any{
		"season_id":       seasonID,
		"division_id":     divID,
		"black_player_id": pidA,
		"white_player_id": pidB,
		"server":          "ogs",
	}).RequireStatus(t, 201)

	var m struct {
		ID int64 `json:"id"`
	}
	create.JSON(t, &m)

	do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/accept", tokB, nil).RequireStatus(t, 200)
	do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/start", tokA, nil).RequireStatus(t, 200)
	do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/result", tokA, map[string]any{
		"winner_id": pidA,
	}).RequireStatus(t, 200)

	// A пытается подтвердить свой же результат.
	do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/confirm", tokA, nil).
		RequireStatus(t, 403)
}

func TestMatch_CannotAcceptOwnChallenge(t *testing.T) {
	truncateAll(t)
	tokA, _, pidA, pidB, seasonID, divID := setupMatchScenario(t)

	create := do(t, "POST", "/api/v1/matches", tokA, map[string]any{
		"season_id":       seasonID,
		"division_id":     divID,
		"black_player_id": pidA,
		"white_player_id": pidB,
		"server":          "ogs",
	}).RequireStatus(t, 201)

	var m struct {
		ID int64 `json:"id"`
	}
	create.JSON(t, &m)

	// Создатель пытается принять свой же вызов.
	do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/accept", tokA, nil).
		RequireStatus(t, 403)
}

func TestMatch_ListMine(t *testing.T) {
	truncateAll(t)
	tokA, tokB, pidA, pidB, seasonID, divID := setupMatchScenario(t)

	// Две партии.
	for i := 0; i < 2; i++ {
		do(t, "POST", "/api/v1/matches", tokA, map[string]any{
			"season_id":       seasonID,
			"division_id":     divID,
			"black_player_id": pidA,
			"white_player_id": pidB,
			"server":          "ogs",
		}).RequireStatus(t, 201)
	}

	listA := do(t, "GET", "/api/v1/me/matches", tokA, nil).RequireStatus(t, 200)
	var matchesA []map[string]any
	listA.JSON(t, &matchesA)
	require.Len(t, matchesA, 2)

	listB := do(t, "GET", "/api/v1/me/matches", tokB, nil).RequireStatus(t, 200)
	var matchesB []map[string]any
	listB.JSON(t, &matchesB)
	require.Len(t, matchesB, 2)
}

func TestMatch_CancelByPlayer(t *testing.T) {
	truncateAll(t)
	tokA, _, pidA, pidB, seasonID, divID := setupMatchScenario(t)

	create := do(t, "POST", "/api/v1/matches", tokA, map[string]any{
		"season_id":       seasonID,
		"division_id":     divID,
		"black_player_id": pidA,
		"white_player_id": pidB,
		"server":          "ogs",
	}).RequireStatus(t, 201)

	var m struct {
		ID int64 `json:"id"`
	}
	create.JSON(t, &m)

	cancel := do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/cancel", tokA, map[string]any{
		"reason": "не смог",
	}).RequireStatus(t, 200)

	var cancelled struct {
		Status       string `json:"status"`
		CancelReason string `json:"cancel_reason"`
	}
	cancel.JSON(t, &cancelled)
	require.Equal(t, "cancelled", cancelled.Status)
	require.Equal(t, "не смог", cancelled.CancelReason)
}

func TestMatch_LimitReached(t *testing.T) {
	truncateAll(t)
	tokA, tokB, pidA, pidB, seasonID, divID := setupMatchScenario(t)

	// n = 4 по умолчанию. Создаём 4 confirmed партии и пробуем пятую.
	for i := 0; i < 4; i++ {
		create := do(t, "POST", "/api/v1/matches", tokA, map[string]any{
			"season_id":       seasonID,
			"division_id":     divID,
			"black_player_id": pidA,
			"white_player_id": pidB,
			"server":          "ogs",
		}).RequireStatus(t, 201)
		var m struct {
			ID int64 `json:"id"`
		}
		create.JSON(t, &m)

		do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/accept", tokB, nil).RequireStatus(t, 200)
		do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/start", tokA, nil).RequireStatus(t, 200)
		do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/result", tokA, map[string]any{
			"winner_id": pidA,
		}).RequireStatus(t, 200)
		do(t, "POST", "/api/v1/matches/"+itoa(m.ID)+"/confirm", tokB, nil).RequireStatus(t, 200)
	}

	// Пятая — лимит исчерпан.
	do(t, "POST", "/api/v1/matches", tokA, map[string]any{
		"season_id":       seasonID,
		"division_id":     divID,
		"black_player_id": pidA,
		"white_player_id": pidB,
		"server":          "ogs",
	}).RequireStatus(t, 409)
}

func TestMatch_DifferentDivisionsForbidden(t *testing.T) {
	truncateAll(t)

	// Два игрока с разными рейтингами → разные дивизионы.
	pidA, tokA := registerAndLogin(t, "alice@example.com", "alice")
	pidB, tokB := registerAndLogin(t, "bob@example.com", "bob")

	seasonID, _, _, _ := seedSeasonWithDivisions(t)

	do(t, "POST", "/api/v1/seasons/"+itoa(seasonID)+"/join", tokA, nil).RequireStatus(t, 200)
	do(t, "POST", "/api/v1/seasons/"+itoa(seasonID)+"/join", tokB, nil).RequireStatus(t, 200)

	// Оба попадут в B (рейтинг 0). Но мы укажем division_id дивизиона S —
	// игроки там не находятся, use case должен отказать.
	listDivs := do(t, "GET", "/api/v1/divisions/1", "", nil).RequireStatus(t, 200)
	var divS struct {
		ID int64 `json:"id"`
	}
	listDivs.JSON(t, &divS)

	do(t, "POST", "/api/v1/matches", tokA, map[string]any{
		"season_id":       seasonID,
		"division_id":     divS.ID, // S, но игроки в B
		"black_player_id": pidA,
		"white_player_id": pidB,
		"server":          "ogs",
	}).RequireStatus(t, 400)
}
