package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/application/usecases/club"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	"github.com/J0hnLenin/67League/internal/delivery/http/middleware"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type ClubHandler struct {
	clubs      ports.ClubRepository
	players    ports.PlayerRepository
	clubBadges ports.ClubBadgeRepository
	badges     ports.BadgeRepository
	createUC   *club.CreateClub
	joinUC     *club.JoinClub
}

func NewClubHandler(
	clubs ports.ClubRepository,
	players ports.PlayerRepository,
	clubBadges ports.ClubBadgeRepository,
	badges ports.BadgeRepository,
	createUC *club.CreateClub,
	joinUC *club.JoinClub,
) *ClubHandler {
	return &ClubHandler{
		clubs:      clubs,
		players:    players,
		clubBadges: clubBadges,
		badges:     badges,
		createUC:   createUC,
		joinUC:     joinUC,
	}
}

func (h *ClubHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.clubs.List(r.Context(), ports.ClubFilter{Pagination: ports.DefaultPagination()})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]appdto.ClubResponse, 0, len(items))
	for _, c := range items {
		n, _ := h.players.CountByClub(r.Context(), c.ID)
		out = append(out, appdto.ClubToResponse(c, n))
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func (h *ClubHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	c, err := h.clubs.GetByID(r.Context(), shared.ClubID(id))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	n, _ := h.players.CountByClub(r.Context(), c.ID)
	httputil.WriteJSON(w, http.StatusOK, appdto.ClubToResponse(c, n))
}

func (h *ClubHandler) Create(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}
	var req appdto.CreateClubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	out, err := h.createUC.Execute(r.Context(), dto.CreateClubInput{
		Name:       req.Name,
		City:       req.City,
		University: req.University,
		CreatorID:  pid,
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, appdto.ClubToResponse(out.Club, 1))
}

func (h *ClubHandler) Join(w http.ResponseWriter, r *http.Request) {
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
	out, err := h.joinUC.Execute(r.Context(), dto.JoinClubInput{
		PlayerID: pid,
		ClubID:   shared.ClubID(id),
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.JoinClubResponse{
		ClubID:           id,
		MemberCount:      out.MemberCount,
		FullBadgeAwarded: out.FullBadgeAwarded,
	})
}

func (h *ClubHandler) Leave(w http.ResponseWriter, r *http.Request) {
	httputil.WriteError(w, errors.New("not implemented"))
}

func (h *ClubHandler) Badges(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	list, err := h.clubBadges.ListByClub(r.Context(), shared.ClubID(id))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]appdto.PlayerBadgeResponse, 0, len(list))
	for _, cb := range list {
		b, err := h.badges.GetByID(r.Context(), cb.BadgeID)
		if err != nil {
			continue
		}
		item := appdto.PlayerBadgeResponse{
			ID:        int64(cb.ID),
			BadgeID:   int64(cb.BadgeID),
			Code:      b.Code,
			Icon:      b.Icon,
			Name:      b.Name,
			AwardedAt: cb.AwardedAt.UTC().Format(time.RFC3339),
			IsActive:  true,
		}
		out = append(out, item)
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}
