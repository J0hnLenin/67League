package middleware

import (
	"context"
	"strconv"

	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type ctxKey int

const (
	ctxKeyAuth ctxKey = iota
	ctxKeyRequestID
)

type AuthInfo struct {
	Subject ports.TokenSubject
	UserID  string
}

func WithAuth(ctx context.Context, a AuthInfo) context.Context {
	return context.WithValue(ctx, ctxKeyAuth, a)
}

func AuthFrom(ctx context.Context) (AuthInfo, bool) {
	a, ok := ctx.Value(ctxKeyAuth).(AuthInfo)
	return a, ok
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, id)
}

func RequestIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKeyRequestID).(string)
	return id, ok
}

// PlayerID возвращает PlayerID из контекста, если субъект — игрок.
func PlayerID(ctx context.Context) (shared.PlayerID, bool) {
	a, ok := AuthFrom(ctx)
	if !ok || a.Subject != ports.SubjectPlayer {
		return 0, false
	}
	n, err := strconv.ParseInt(a.UserID, 10, 64)
	if err != nil {
		return 0, false
	}
	return shared.PlayerID(n), true
}

// AdminID возвращает AdminID из контекста, если субъект — админ.
func AdminID(ctx context.Context) (shared.AdminID, bool) {
	a, ok := AuthFrom(ctx)
	if !ok || a.Subject != ports.SubjectAdmin {
		return 0, false
	}
	n, err := strconv.ParseInt(a.UserID, 10, 64)
	if err != nil {
		return 0, false
	}
	return shared.AdminID(n), true
}
