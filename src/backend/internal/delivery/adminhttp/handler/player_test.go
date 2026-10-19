package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

func TestAdminPlayerHandler_List(t *testing.T) {
	repo := new(mocks.PlayerRepository)
	matchRepo := new(mocks.MatchRepository)
	h := NewAdminPlayerHandler(repo, matchRepo)

	repo.On("List", mock.Anything, mock.Anything).Return([]*player.Player{
		{ID: 1, FullName: "Alice", Nickname: "alice", City: "Moscow", Status: player.StatusActive},
		{ID: 2, FullName: "Bob", Nickname: "bob", City: "SPb", Status: player.StatusBanned},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/players", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var out []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 2)
	require.Equal(t, float64(1), out[0]["id"])
	require.Equal(t, "alice", out[0]["nickname"])
	require.Equal(t, "banned", out[1]["status"])
}

func TestAdminPlayerHandler_GetByID(t *testing.T) {
	repo := new(mocks.PlayerRepository)
	matchRepo := new(mocks.MatchRepository)
	h := NewAdminPlayerHandler(repo, matchRepo)

	repo.On("GetByID", mock.Anything, shared.PlayerID(1)).Return(
		&player.Player{ID: 1, FullName: "Alice", Nickname: "alice", City: "Moscow", Status: player.StatusActive},
		nil,
	)

	r := chi.NewRouter()
	r.Get("/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Equal(t, float64(1), out["id"])
	require.Equal(t, "alice", out["nickname"])
}

func TestAdminPlayerHandler_Update(t *testing.T) {
	repo := new(mocks.PlayerRepository)
	matchRepo := new(mocks.MatchRepository)
	h := NewAdminPlayerHandler(repo, matchRepo)

	p := &player.Player{ID: 1, FullName: "Alice", Nickname: "alice", City: "Moscow", Status: player.StatusActive}
	repo.On("GetByID", mock.Anything, shared.PlayerID(1)).Return(p, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"full_name": "Alice Updated",
		"city":      "Kazan",
		"rating":    1500,
	})

	r := chi.NewRouter()
	r.Patch("/{id}", h.Update)

	req := httptest.NewRequest(http.MethodPatch, "/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Equal(t, "Alice Updated", out["full_name"])
	require.Equal(t, "Kazan", out["city"])
	require.Equal(t, float64(1500), out["rating"])
}

func TestAdminPlayerHandler_Ban(t *testing.T) {
	repo := new(mocks.PlayerRepository)
	matchRepo := new(mocks.MatchRepository)
	h := NewAdminPlayerHandler(repo, matchRepo)

	p := &player.Player{ID: 1, FullName: "Alice", Nickname: "alice", Status: player.StatusActive}
	repo.On("GetByID", mock.Anything, shared.PlayerID(1)).Return(p, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	r := chi.NewRouter()
	r.Post("/{id}/ban", h.Ban)

	req := httptest.NewRequest(http.MethodPost, "/1/ban", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Equal(t, "banned", out["status"])
}

func TestAdminPlayerHandler_Unban(t *testing.T) {
	repo := new(mocks.PlayerRepository)
	matchRepo := new(mocks.MatchRepository)
	h := NewAdminPlayerHandler(repo, matchRepo)

	p := &player.Player{ID: 1, FullName: "Alice", Nickname: "alice", Status: player.StatusBanned}
	repo.On("GetByID", mock.Anything, shared.PlayerID(1)).Return(p, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	r := chi.NewRouter()
	r.Post("/{id}/unban", h.Unban)

	req := httptest.NewRequest(http.MethodPost, "/1/unban", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Equal(t, "active", out["status"])
}
