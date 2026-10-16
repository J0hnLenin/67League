package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateClub(t *testing.T) {
	truncateAll(t)
	_, token := registerAndLogin(t, "alice@example.com", "alice")

	resp := do(t, "POST", "/api/v1/clubs", token, map[string]any{
		"name": "Shibuya",
		"city": "Tokyo",
	}).RequireStatus(t, 201)

	var created struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		Members int    `json:"members"`
	}
	resp.JSON(t, &created)
	require.NotZero(t, created.ID)
	require.Equal(t, "Shibuya", created.Name)
	require.Equal(t, 1, created.Members)

	// Клуб доступен публично.
	list := do(t, "GET", "/api/v1/clubs", "", nil).RequireStatus(t, 200)
	var clubs []map[string]any
	list.JSON(t, &clubs)
	require.Len(t, clubs, 1)
	require.Equal(t, "Shibuya", clubs[0]["name"])
}

func TestCreateClub_DuplicateName(t *testing.T) {
	truncateAll(t)
	_, token := registerAndLogin(t, "alice@example.com", "alice")

	do(t, "POST", "/api/v1/clubs", token, map[string]any{
		"name": "Shibuya",
	}).RequireStatus(t, 201)

	// Второй игрок пытается создать клуб с тем же именем.
	_, token2 := registerAndLogin(t, "bob@example.com", "bob")
	do(t, "POST", "/api/v1/clubs", token2, map[string]any{
		"name": "Shibuya",
	}).RequireStatus(t, 409)
}

func TestJoinClub(t *testing.T) {
	truncateAll(t)
	_, token1 := registerAndLogin(t, "alice@example.com", "alice")
	_, token2 := registerAndLogin(t, "bob@example.com", "bob")

	create := do(t, "POST", "/api/v1/clubs", token1, map[string]any{
		"name": "Shibuya",
	}).RequireStatus(t, 201)

	var c struct {
		ID int64 `json:"id"`
	}
	create.JSON(t, &c)

	join := do(t, "POST", "/api/v1/clubs/"+itoa(c.ID)+"/join", token2, nil).
		RequireStatus(t, 200)

	var joined struct {
		ClubID      int64 `json:"club_id"`
		MemberCount int   `json:"member_count"`
	}
	join.JSON(t, &joined)
	require.Equal(t, c.ID, joined.ClubID)
	require.Equal(t, 2, joined.MemberCount)

	// Публично видно 2 участника.
	get := do(t, "GET", "/api/v1/clubs/"+itoa(c.ID), "", nil).RequireStatus(t, 200)
	var club struct {
		Members int `json:"members"`
	}
	get.JSON(t, &club)
	require.Equal(t, 2, club.Members)
}

func TestJoinClub_AlreadyMember(t *testing.T) {
	truncateAll(t)
	_, token := registerAndLogin(t, "alice@example.com", "alice")

	create := do(t, "POST", "/api/v1/clubs", token, map[string]any{
		"name": "Shibuya",
	}).RequireStatus(t, 201)

	var c struct {
		ID int64 `json:"id"`
	}
	create.JSON(t, &c)

	// Создатель уже в клубе — повторный вход даёт конфликт.
	do(t, "POST", "/api/v1/clubs/"+itoa(c.ID)+"/join", token, nil).
		RequireStatus(t, 409)
}

func TestJoinClub_NotFound(t *testing.T) {
	truncateAll(t)
	_, token := registerAndLogin(t, "alice@example.com", "alice")

	do(t, "POST", "/api/v1/clubs/99999/join", token, nil).
		RequireStatus(t, 404)
}
