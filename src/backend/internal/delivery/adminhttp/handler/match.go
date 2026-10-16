package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	appdtoapp "github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/usecases/match"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type AdminMatchHandler struct {
	cancel *match.CancelMatch
}

func NewAdminMatchHandler(cancel *match.CancelMatch) *AdminMatchHandler {
	return &AdminMatchHandler{cancel: cancel}
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
