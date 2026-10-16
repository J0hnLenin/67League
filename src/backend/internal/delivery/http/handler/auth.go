package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/usecases/player"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"

	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type AuthHandler struct {
	register *player.RegisterPlayer
	login    *player.LoginPlayer
	logger   *slog.Logger
}

func NewAuthHandler(
	register *player.RegisterPlayer,
	login *player.LoginPlayer,
	logger *slog.Logger,
) *AuthHandler {
	return &AuthHandler{register: register, login: login, logger: logger}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req appdto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	email, err := shared.NewEmail(req.Email)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	nickname, err := shared.NewNickname(req.Nickname)
	if err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	var clubID *shared.ClubID
	if req.ClubID != nil {
		id := shared.ClubID(*req.ClubID)
		clubID = &id
	}

	out, err := h.register.Execute(r.Context(), dto.RegisterPlayerInput{
		FullName:  req.FullName,
		Email:     email,
		Password:  req.Password,
		Nickname:  nickname,
		City:      req.City,
		KGSNick:   req.KGSNick,
		OGSNick:   req.OGSNick,
		FoxNick:   req.FoxNick,
		RFGID:     req.RFGID,
		SputnikID: req.SputnikID,
		ClubID:    clubID,
	})
	if err != nil {
		httputil.LogError(h.logger, err, "path", "auth.register")
		httputil.WriteError(w, err)
		return
	}

	resp := appdto.RegisterResponse{PlayerID: int64(out.Player.ID)}
	if out.Division != nil {
		d := int64(*out.Division)
		resp.DivisionID = &d
	}
	httputil.WriteJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req appdto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	email, err := shared.NewEmail(req.Email)
	if err != nil {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}

	out, err := h.login.Execute(r.Context(), dto.LoginPlayerInput{
		Email:    email,
		Password: req.Password,
	})
	if err != nil {
		if !errors.Is(err, shared.ErrUnauthorized) {
			httputil.LogError(h.logger, err, "path", "auth.login")
		}
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, appdto.LoginResponse{
		PlayerID: int64(out.Player.ID),
		Nickname: string(out.Player.Nickname),
		Token:    out.Token,
	})
}
