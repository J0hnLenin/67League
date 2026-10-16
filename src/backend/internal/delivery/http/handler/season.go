package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	usecaseseason "github.com/J0hnLenin/67League/internal/application/usecases/season"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	"github.com/J0hnLenin/67League/internal/delivery/http/middleware"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type SeasonHandler struct {
	seasons ports.SeasonRepository
	joinUC  *usecaseseason.JoinSeason
}

func NewSeasonHandler(
	seasons ports.SeasonRepository,
	joinUC *usecaseseason.JoinSeason,
) *SeasonHandler {
	return &SeasonHandler{seasons: seasons, joinUC: joinUC}
}

func (h *SeasonHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.seasons.List(r.Context(), ports.SeasonFilter{Pagination: ports.DefaultPagination()})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]appdto.SeasonResponse, 0, len(list))
	for _, s := range list {
		out = append(out, seasonToResponse(s))
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func (h *SeasonHandler) Current(w http.ResponseWriter, r *http.Request) {
	s, err := h.seasons.GetCurrent(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, seasonToResponse(s))
}

func (h *SeasonHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	s, err := h.seasons.GetByID(r.Context(), shared.SeasonID(id))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, seasonToResponse(s))
}

func (h *SeasonHandler) Join(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	out, err := h.joinUC.Execute(r.Context(), dto.JoinSeasonInput{
		PlayerID: pid,
		SeasonID: shared.SeasonID(id),
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.JoinSeasonResponse{
		DivisionID: int64(out.Division.ID),
		MemberID:   int64(out.Member.ID),
	})
}

func seasonToResponse(s *domseason.Season) appdto.SeasonResponse {
	return appdto.SeasonResponse{
		ID:        int64(s.ID),
		Number:    s.Number,
		Title:     s.Title,
		StartDate: s.StartDate.UTC().Format(time.RFC3339),
		EndDate:   s.EndDate.UTC().Format(time.RFC3339),
		Status:    string(s.Status),
	}
}
