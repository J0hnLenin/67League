package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/application/usecases/match"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	"github.com/J0hnLenin/67League/internal/delivery/http/middleware"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type MatchHandler struct {
	matches ports.MatchRepository
	events  ports.MatchEventRepository

	createUC  *match.CreateMatch
	acceptUC  *match.AcceptMatch
	startUC   *match.StartMatch
	submitUC  *match.SubmitResult
	confirmUC *match.ConfirmResult
	cancelUC  *match.CancelMatch
}

func NewMatchHandler(
	matches ports.MatchRepository,
	events ports.MatchEventRepository,
	createUC *match.CreateMatch,
	acceptUC *match.AcceptMatch,
	startUC *match.StartMatch,
	submitUC *match.SubmitResult,
	confirmUC *match.ConfirmResult,
	cancelUC *match.CancelMatch,
) *MatchHandler {
	return &MatchHandler{
		matches:   matches,
		events:    events,
		createUC:  createUC,
		acceptUC:  acceptUC,
		startUC:   startUC,
		submitUC:  submitUC,
		confirmUC: confirmUC,
		cancelUC:  cancelUC,
	}
}

func (h *MatchHandler) Create(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}

	var req appdto.CreateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}

	out, err := h.createUC.Execute(r.Context(), dto.CreateMatchInput{
		SeasonID:      shared.SeasonID(req.SeasonID),
		DivisionID:    shared.DivisionID(req.DivisionID),
		BlackPlayerID: shared.PlayerID(req.BlackPlayerID),
		WhitePlayerID: shared.PlayerID(req.WhitePlayerID),
		CreatedBy:     pid,
		Server:        dommatch.Server(req.Server),
		TimeControl:   req.TimeControl,
		Handicap:      req.Handicap,
		Komi:          req.Komi,
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, matchToResponse(out.Match))
}

func (h *MatchHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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
	httputil.WriteJSON(w, http.StatusOK, matchToResponse(m))
}

func (h *MatchHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}

	f := ports.MatchFilter{Pagination: ports.DefaultPagination()}
	if s := r.URL.Query().Get("status"); s != "" {
		st := dommatch.Status(s)
		if !st.Valid() {
			httputil.WriteError(w, shared.ErrValidation)
			return
		}
		f.Status = &st
	}

	list, err := h.matches.ListByPlayer(r.Context(), pid, f)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]appdto.MatchResponse, 0, len(list))
	for _, m := range list {
		out = append(out, matchToResponse(m))
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func (h *MatchHandler) Accept(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	out, err := h.acceptUC.Execute(r.Context(), dto.AcceptMatchInput{
		MatchID: shared.MatchID(id),
		ActorID: pid,
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, matchToResponse(out.Match))
}

func (h *MatchHandler) Start(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	out, err := h.startUC.Execute(r.Context(), dto.StartMatchInput{
		MatchID: shared.MatchID(id),
		ActorID: pid,
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, matchToResponse(out.Match))
}

func (h *MatchHandler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}

	var req appdto.SubmitResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}

	out, err := h.submitUC.Execute(r.Context(), dto.SubmitResultInput{
		MatchID:   shared.MatchID(id),
		ActorID:   pid,
		WinnerID:  shared.PlayerID(req.WinnerID),
		ScoreDiff: req.ScoreDiff,
		SGFURL:    req.SGFURL,
		Link:      req.Link,
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, matchToResponse(out.Match))
}

func (h *MatchHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	out, err := h.confirmUC.Execute(r.Context(), dto.ConfirmResultInput{
		MatchID: shared.MatchID(id),
		ActorID: pid,
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, matchToResponse(out.Match))
}

func (h *MatchHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	pid, ok := middleware.PlayerID(r.Context())
	if !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}

	var req appdto.CancelMatchRequest
	// тело необязательное — если пусто, req останется нулевым
	_ = json.NewDecoder(r.Body).Decode(&req)

	out, err := h.cancelUC.Execute(r.Context(), dto.CancelMatchInput{
		MatchID: shared.MatchID(id),
		ActorID: pid,
		IsAdmin: false,
		Reason:  req.Reason,
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, matchToResponse(out.Match))
}

// ---- helpers --------------------------------------------------------------

func matchToResponse(m *dommatch.Match) appdto.MatchResponse {
	resp := appdto.MatchResponse{
		ID:            int64(m.ID),
		SeasonID:      int64(m.SeasonID),
		DivisionID:    int64(m.DivisionID),
		BlackPlayerID: int64(m.BlackPlayerID),
		WhitePlayerID: int64(m.WhitePlayerID),
		CreatedBy:     int64(m.CreatedBy),
		Status:        string(m.Status),
		Server:        string(m.Server),
		TimeControl:   m.TimeControl,
		Handicap:      m.Handicap,
		Komi:          m.Komi,
		SGFURL:        m.SGFURL,
		Link:          m.ExternalLink,
		CancelReason:  m.CancelReason,
		CreatedAt:     m.CreatedAt.UTC().Format(time.RFC3339),
	}
	if m.WinnerID != nil {
		id := int64(*m.WinnerID)
		resp.WinnerID = &id
	}
	if m.ScoreDiff != nil {
		resp.ScoreDiff = m.ScoreDiff
	}
	if m.AcceptedAt != nil {
		s := m.AcceptedAt.UTC().Format(time.RFC3339)
		resp.AcceptedAt = &s
	}
	if m.StartedAt != nil {
		s := m.StartedAt.UTC().Format(time.RFC3339)
		resp.StartedAt = &s
	}
	if m.FinishedAt != nil {
		s := m.FinishedAt.UTC().Format(time.RFC3339)
		resp.FinishedAt = &s
	}
	if m.ConfirmedAt != nil {
		s := m.ConfirmedAt.UTC().Format(time.RFC3339)
		resp.ConfirmedAt = &s
	}
	if m.CancelledAt != nil {
		s := m.CancelledAt.UTC().Format(time.RFC3339)
		resp.CancelledAt = &s
	}
	return resp
}

// parseID объявлен в handler/auth.go, здесь не дублируем.
var _ = strconv.ParseInt
