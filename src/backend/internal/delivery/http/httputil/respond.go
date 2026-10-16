package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func WriteError(w http.ResponseWriter, err error) {
	status, code := mapError(err)
	body := map[string]any{
		"error": code,
	}
	if code == "internal" {
		body["message"] = "internal error"
	} else {
		body["message"] = err.Error()
	}
	WriteJSON(w, status, body)
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, shared.ErrValidation):
		return http.StatusBadRequest, "validation"
	case errors.Is(err, shared.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, shared.ErrAlreadyExists):
		return http.StatusConflict, "already_exists"
	case errors.Is(err, shared.ErrConflict):
		return http.StatusConflict, "conflict"
	case errors.Is(err, shared.ErrForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, shared.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, shared.ErrInvalidState):
		return http.StatusConflict, "invalid_state"
	case errors.Is(err, shared.ErrLimitExceeded):
		return http.StatusConflict, "limit_exceeded"
	default:
		return http.StatusInternalServerError, "internal"
	}
}

func LogError(logger *slog.Logger, err error, ctx ...any) {
	logger.Error("request failed", append([]any{"err", err}, ctx...)...)
}
