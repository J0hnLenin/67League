package middleware

import (
	"net/http"
	"strings"

	"github.com/J0hnLenin/67League/internal/application/ports"
)

// Auth извлекает Bearer-токен, парсит его и кладёт AuthInfo в контекст.
// Если токена нет или он невалиден — возвращает 401.
func Auth(tokens ports.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := tokens.Parse(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := WithAuth(r.Context(), AuthInfo{
				Subject: claims.Subject,
				UserID:  claims.UserID,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePlayer пропускает дальше только игрока.
// Ставится после Auth.
func RequirePlayer() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := PlayerID(r.Context()); !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin пропускает дальше только админа.
// Ставится после Auth.
func RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := AdminID(r.Context()); !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
