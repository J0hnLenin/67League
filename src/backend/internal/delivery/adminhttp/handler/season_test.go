package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

func TestAdminSeasonHandler_List(t *testing.T) {
	seasonRepo := new(mocks.SeasonRepository)
	divRepo := new(mocks.DivisionRepository)
	h := NewAdminSeasonHandler(seasonRepo, divRepo, nil, nil, nil, nil, nil)

	seasonRepo.On("List", mock.Anything, mock.Anything).Return([]*season.Season{
		{ID: 1, Number: 1, Title: "Season 1", StartDate: time.Now(), EndDate: time.Now().Add(30 * 24 * time.Hour), Status: season.StatusActive},
		{ID: 2, Number: 2, Title: "Season 2", StartDate: time.Now(), EndDate: time.Now().Add(30 * 24 * time.Hour), Status: season.StatusAnnounced},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/seasons", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var out []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 2)
	require.Equal(t, "Season 1", out[0]["title"])
	require.Equal(t, "active", out[0]["status"])
	require.Equal(t, "announced", out[1]["status"])
}
