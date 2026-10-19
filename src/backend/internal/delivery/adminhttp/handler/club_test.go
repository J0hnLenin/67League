package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

func TestAdminClubHandler_List(t *testing.T) {
	clubRepo := new(mocks.ClubRepository)
	playerRepo := new(mocks.PlayerRepository)
	h := NewAdminClubHandler(clubRepo, playerRepo)

	clubRepo.On("List", mock.Anything, mock.Anything).Return([]*club.Club{
		{ID: 1, Name: "Shibuya", City: "Tokyo", Status: club.StatusActive},
		{ID: 2, Name: "Akihabara", City: "Tokyo", Status: club.StatusArchived},
	}, nil)
	playerRepo.On("CountByClub", mock.Anything, mock.Anything).Return(5, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/clubs", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var out []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 2)
	require.Equal(t, "Shibuya", out[0]["name"])
	require.Equal(t, "Tokyo", out[0]["city"])
	require.Equal(t, float64(5), out[0]["members"])
	require.Equal(t, "active", out[0]["status"])
	require.Equal(t, "archived", out[1]["status"])
}
