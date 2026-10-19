package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	appdtoapp "github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/application/usecases/match"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type AdminMatchHandler struct {
	matches ports.MatchRepository
	cancel  *match.CancelMatch
}

func NewAdminMatchHandler(matches ports.MatchRepository, cancel *match.CancelMatch) *AdminMatchHandler {
	return &AdminMatchHandler{matches: matches, cancel: cancel}
}

func (h *AdminMatchHandler) List(w http.ResponseWriter, r *http.Request) {
	f := ports.MatchFilter{Pagination: ports.DefaultPagination()}
	list, err := h.matches.List(r.Context(), f)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.MatchesToResponse(list))
}

func (h *AdminMatchHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	m, err := h.matches.GetByID(r.Context(), shared.MatchID(id))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.MatchToResponse(m))
}

func (h *AdminMatchHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	var body appdto.CancelMatchRequest
	_ = json.NewDecoder(r.Body).Decode(&body)

	out, err := h.cancel.Execute(r.Context(), appdtoapp.CancelMatchInput{
		MatchID: shared.MatchID(id),
		ActorID: 0,
		IsAdmin: true,
		Reason:  body.Reason,
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.MatchToResponse(out.Match))
}
