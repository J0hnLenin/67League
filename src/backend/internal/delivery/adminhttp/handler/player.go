package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/J0hnLenin/67League/internal/application/ports"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type AdminPlayerHandler struct {
	players ports.PlayerRepository
}

func NewAdminPlayerHandler(players ports.PlayerRepository) *AdminPlayerHandler {
	return &AdminPlayerHandler{players: players}
}

func (h *AdminPlayerHandler) List(w http.ResponseWriter, r *http.Request) {
	f := ports.PlayerFilter{Pagination: ports.DefaultPagination()}
	if s := r.URL.Query().Get("status"); s != "" {
		st := domplayer.Status(s)
		f.Status = &st
	}
	list, err := h.players.List(r.Context(), f)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.PlayersToResponse(list))
}

func (h *AdminPlayerHandler) Ban(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, func(p *domplayer.Player) error {
		return p.Ban(timeNowUTC())
	})
}

func (h *AdminPlayerHandler) Unban(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, func(p *domplayer.Player) error {
		return p.Unban(timeNowUTC())
	})
}

func (h *AdminPlayerHandler) setStatus(
	w http.ResponseWriter,
	r *http.Request,
	fn func(*domplayer.Player) error,
) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	p, err := h.players.GetByID(r.Context(), shared.PlayerID(id))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	if err := fn(p); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if err := h.players.Update(r.Context(), p); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"id":     int64(p.ID),
		"status": string(p.Status),
	})
}
