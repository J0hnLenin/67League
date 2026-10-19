package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/J0hnLenin/67League/internal/application/ports"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	"github.com/J0hnLenin/67League/internal/delivery/http/middleware"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type PlayerHandler struct {
	players      ports.PlayerRepository
	badges       ports.BadgeRepository
	playerBadges ports.PlayerBadgeRepository
	titles       ports.TitleRepository
	stones       ports.StoneTransactionRepository
}

func NewPlayerHandler(
	players ports.PlayerRepository,
	badges ports.BadgeRepository,
	playerBadges ports.PlayerBadgeRepository,
	titles ports.TitleRepository,
	stones ports.StoneTransactionRepository,
) *PlayerHandler {
	return &PlayerHandler{
		players:      players,
		badges:       badges,
		playerBadges: playerBadges,
		titles:       titles,
		stones:       stones,
	}
}

func (h *PlayerHandler) Me(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}
	h.writePlayer(w, r, pid)
}

func (h *PlayerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	h.writePlayer(w, r, shared.PlayerID(id))
}

func (h *PlayerHandler) writePlayer(w http.ResponseWriter, r *http.Request, id shared.PlayerID) {
	p, err := h.players.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.PlayerToResponse(p))
}

func (h *PlayerHandler) Badges(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}

	list, err := h.playerBadges.ListByPlayer(r.Context(), pid)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	out := make([]appdto.PlayerBadgeResponse, 0, len(list))
	for _, pb := range list {
		b, err := h.badges.GetByID(r.Context(), pb.BadgeID)
		if err != nil {
			continue
		}
		item := appdto.PlayerBadgeResponse{
			ID:        int64(pb.ID),
			BadgeID:   int64(pb.BadgeID),
			Code:      b.Code,
			Icon:      b.Icon,
			Name:      b.Name,
			AwardedAt: pb.AwardedAt.UTC().Format(time.RFC3339),
			IsActive:  pb.IsActive,
		}
		if pb.SeasonID != nil {
			sid := int64(*pb.SeasonID)
			item.SeasonID = &sid
		}
		out = append(out, item)
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func (h *PlayerHandler) Titles(w http.ResponseWriter, r *http.Request) {
	list, err := h.titles.List(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]appdto.TitleResponse, 0, len(list))
	for _, t := range list {
		item := appdto.TitleResponse{ID: int64(t.ID), Code: t.Code, Name: t.Name}
		if t.BadgeID != nil {
			id := int64(*t.BadgeID)
			item.BadgeID = &id
		}
		out = append(out, item)
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func (h *PlayerHandler) StoneHistory(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}
	list, err := h.stones.ListByPlayer(r.Context(), pid, ports.DefaultPagination())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]appdto.StoneTransactionResponse, 0, len(list))
	for _, tx := range list {
		out = append(out, appdto.StoneTransactionResponse{
			ID:        int64(tx.ID),
			Amount:    float64(tx.Amount),
			Reason:    tx.Reason,
			CreatedAt: tx.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func (h *PlayerHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		httputil.WriteJSON(w, http.StatusOK, []appdto.PlayerSearchResponse{})
		return
	}

	list, err := h.players.Search(r.Context(), query, 10)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	out := make([]appdto.PlayerSearchResponse, 0, len(list))
	for _, p := range list {
		out = append(out, appdto.PlayerSearchResponse{
			ID:       int64(p.ID),
			FullName: p.FullName,
			Nickname: string(p.Nickname),
		})
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}
