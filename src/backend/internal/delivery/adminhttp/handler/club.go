package handler

import (
	"net/http"

	"github.com/J0hnLenin/67League/internal/application/ports"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
)

type AdminClubHandler struct {
	clubs   ports.ClubRepository
	players ports.PlayerRepository
}

func NewAdminClubHandler(clubs ports.ClubRepository, players ports.PlayerRepository) *AdminClubHandler {
	return &AdminClubHandler{clubs: clubs, players: players}
}

func (h *AdminClubHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.clubs.List(r.Context(), ports.ClubFilter{Pagination: ports.DefaultPagination()})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]appdto.ClubResponse, 0, len(list))
	for _, c := range list {
		n, _ := h.players.CountByClub(r.Context(), c.ID)
		out = append(out, appdto.ClubToResponse(c, n))
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}
