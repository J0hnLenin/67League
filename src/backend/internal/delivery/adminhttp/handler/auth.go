package handler

import (
	"encoding/json"
	"net/http"

	"github.com/J0hnLenin/67League/internal/application/ports"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type AdminAuthHandler struct {
	admins ports.AdminRepository
	hasher ports.PasswordHasher
	tokens ports.TokenIssuer
}

func NewAdminAuthHandler(
	admins ports.AdminRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenIssuer,
) *AdminAuthHandler {
	return &AdminAuthHandler{admins: admins, hasher: hasher, tokens: tokens}
}

func (h *AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req appdto.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}

	a, err := h.admins.GetByLogin(r.Context(), req.Login)
	if err != nil {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}

	ok, err := h.hasher.Verify(req.Password, a.PasswordHash)
	if err != nil || !ok {
		httputil.WriteError(w, shared.ErrUnauthorized)
		return
	}

	token, err := h.tokens.IssueAdmin(a.ID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, appdto.AdminLoginResponse{
		AdminID: int64(a.ID),
		Token:   token,
	})
}
