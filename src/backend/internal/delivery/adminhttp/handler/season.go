package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	appdtoapp "github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	badgeuc "github.com/J0hnLenin/67League/internal/application/usecases/badge"
	seasonuc "github.com/J0hnLenin/67League/internal/application/usecases/season"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type AdminSeasonHandler struct {
	seasons   ports.SeasonRepository
	divisions ports.DivisionRepository

	calculate *seasonuc.CalculateSeasonResults
	rotate    *seasonuc.ApplyRotation

	awardPersonal   *badgeuc.AwardPersonalBadges
	awardSeason     *badgeuc.AwardSeasonBadges
	awardSuperTitle *badgeuc.AwardSuperTitle
}

func (h *AdminSeasonHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.seasons.List(r.Context(), ports.SeasonFilter{Pagination: ports.DefaultPagination()})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]appdto.SeasonResponse, 0, len(list))
	for _, s := range list {
		out = append(out, appdto.SeasonToResponse(s))
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func NewAdminSeasonHandler(
	seasons ports.SeasonRepository,
	divisions ports.DivisionRepository,
	calculate *seasonuc.CalculateSeasonResults,
	rotate *seasonuc.ApplyRotation,
	awardPersonal *badgeuc.AwardPersonalBadges,
	awardSeason *badgeuc.AwardSeasonBadges,
	awardSuperTitle *badgeuc.AwardSuperTitle,
) *AdminSeasonHandler {
	return &AdminSeasonHandler{
		seasons:         seasons,
		divisions:       divisions,
		calculate:       calculate,
		rotate:          rotate,
		awardPersonal:   awardPersonal,
		awardSeason:     awardSeason,
		awardSuperTitle: awardSuperTitle,
	}
}

// ---- Переходы состояний -----------------------------------------------

func (h *AdminSeasonHandler) StartRegistration(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, func(s *domseason.Season, now time.Time) error {
		return s.StartRegistration(now)
	})
}

func (h *AdminSeasonHandler) Start(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, func(s *domseason.Season, now time.Time) error {
		return s.Start(now)
	})
}

func (h *AdminSeasonHandler) Close(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, func(s *domseason.Season, now time.Time) error {
		return s.Close(now)
	})
}

func (h *AdminSeasonHandler) Finish(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, func(s *domseason.Season, now time.Time) error {
		return s.Finish(now)
	})
}

func (h *AdminSeasonHandler) Archive(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, func(s *domseason.Season, now time.Time) error {
		return s.Archive(now)
	})
}

func (h *AdminSeasonHandler) transition(
	w http.ResponseWriter,
	r *http.Request,
	fn func(*domseason.Season, time.Time) error,
) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	s, err := h.seasons.GetByID(r.Context(), shared.SeasonID(id))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	now := time.Now().UTC()
	if err := fn(s, now); err != nil {
		httputil.WriteError(w, err)
		return
	}
	s.UpdatedAt = now
	if err := h.seasons.Update(r.Context(), s); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.SeasonToResponse(s))
}

// ---- Создание сезона --------------------------------------------------

func (h *AdminSeasonHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req appdto.CreateSeasonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	start, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}

	settings := domseason.DefaultSettings()
	if req.Settings != nil {
		if req.Settings.K1 > 0 {
			settings.K1 = req.Settings.K1
		}
		if req.Settings.K2 > 0 {
			settings.K2 = req.Settings.K2
		}
		if req.Settings.K3 > 0 {
			settings.K3 = req.Settings.K3
		}
		if req.Settings.P > 0 {
			settings.P = req.Settings.P
		}
		if req.Settings.N > 0 {
			settings.N = req.Settings.N
		}
		if req.Settings.T > 0 {
			settings.T = req.Settings.T
		}
	}

	now := time.Now().UTC()
	s, err := domseason.New(req.Number, req.Title, start, end, settings, now)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	if err := h.seasons.Create(r.Context(), s); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, appdto.SeasonToResponse(s))
}

// ---- Итоги ------------------------------------------------------------

func (h *AdminSeasonHandler) CalculateResults(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	out, err := h.calculate.Execute(r.Context(), appdtoapp.CalculateSeasonResultsInput{
		SeasonID: shared.SeasonID(id),
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, calculateResultsToJSON(out))
}

func (h *AdminSeasonHandler) ApplyRotation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NextSeasonID int64 `json:"next_season_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	out, err := h.rotate.Execute(r.Context(), appdtoapp.ApplyRotationInput{
		PreviousSeasonID: shared.SeasonID(id),
		NextSeasonID:     shared.SeasonID(req.NextSeasonID),
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	resp := appdto.RotationResponse{Created: out.Created}
	for _, t := range out.Transfers {
		resp.Transfers = append(resp.Transfers, appdto.PlayerTransferJSON{
			PlayerID:  int64(t.PlayerID),
			FromDivID: int64(t.FromDivID),
			ToDivID:   int64(t.ToDivID),
			Direction: string(t.Direction),
			FromPlace: t.FromPlace,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AdminSeasonHandler) AwardSeasonBadges(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	out, err := h.awardSeason.Execute(r.Context(), appdtoapp.AwardSeasonBadgesInput{
		SeasonID: shared.SeasonID(id),
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	resp := appdto.AwardBadgesResponse{}
	for _, pb := range out.Awarded {
		resp.Awarded = append(resp.Awarded, appdto.PlayerBadgeResponse{
			ID:      int64(pb.ID),
			BadgeID: int64(pb.BadgeID),
		})
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AdminSeasonHandler) AwardSuperTitle(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	out, err := h.awardSuperTitle.Execute(r.Context(), appdtoapp.AwardSuperTitleInput{
		SeasonID: shared.SeasonID(id),
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	resp := appdto.AwardSuperTitleResponse{Awarded: out.Awarded}
	if out.Player != nil {
		pid := int64(*out.Player)
		resp.PlayerID = &pid
	}
	if out.Club != nil {
		cid := int64(*out.Club)
		resp.ClubID = &cid
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// ---- helpers ----------------------------------------------------------

func calculateResultsToJSON(out *appdtoapp.CalculateSeasonResultsOutput) appdto.FinishSeasonResponse {
	resp := appdto.FinishSeasonResponse{
		Season: appdto.SeasonToResponse(out.Season),
	}
	for _, d := range out.Standings {
		row := appdto.DivisionStandingJSON{
			DivisionID: int64(d.Division.ID),
			Code:       d.Division.Code,
		}
		for _, m := range d.Rows {
			row.Rows = append(row.Rows, appdto.MemberStandingJSON{
				PlayerID: int64(m.PlayerID),
				LP:       float64(m.Member.LP),
				Place:    m.Place,
				Excluded: m.Excluded,
			})
		}
		resp.Standings = append(resp.Standings, row)
	}
	for _, c := range out.ClubTable {
		resp.ClubTable = append(resp.ClubTable, appdto.ClubStandingJSON{
			ClubID:  int64(c.ClubID),
			TotalLP: float64(c.TotalLP),
			Members: c.Members,
			Top5Sum: float64(c.Top5Sum),
		})
	}
	if out.BestClub != nil {
		id := int64(out.BestClub.ClubID)
		resp.BestClubID = &id
	}
	return resp
}
