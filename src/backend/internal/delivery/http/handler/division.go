package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/J0hnLenin/67League/internal/application/ports"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type DivisionHandler struct {
	divisions ports.DivisionRepository
	members   ports.DivisionMemberRepository
}

func NewDivisionHandler(
	divisions ports.DivisionRepository,
	members ports.DivisionMemberRepository,
) *DivisionHandler {
	return &DivisionHandler{divisions: divisions, members: members}
}

func (h *DivisionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	d, err := h.divisions.GetByID(r.Context(), shared.DivisionID(id))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.DivisionResponse{
		ID:        int64(d.ID),
		SeasonID:  int64(d.SeasonID),
		Code:      d.Code,
		Alias:     d.Alias,
		RatingMin: int(d.RatingMin),
		RatingMax: int(d.RatingMax),
		Order:     d.Order,
		Status:    string(d.Status),
	})
}

func (h *DivisionHandler) Members(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	list, err := h.members.ListByDivision(r.Context(), shared.DivisionID(id))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]appdto.DivisionMemberResponse, 0, len(list))
	for _, m := range list {
		out = append(out, appdto.DivisionMemberResponse{
			PlayerID:      int64(m.PlayerID),
			DivisionID:    int64(m.DivisionID),
			LP:            float64(m.LP),
			Wins:          m.Wins,
			Losses:        m.Losses,
			MatchesPlayed: m.MatchesPlayed,
			StreakCurrent: m.StreakCurrent,
			StreakMax:     m.StreakMax,
			FinalPlace:    m.FinalPlace,
			Status:        string(m.Status),
		})
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}
