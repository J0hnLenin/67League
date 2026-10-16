package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthz(t *testing.T) {
	truncateAll(t)
	resp := do(t, "GET", "/healthz", "", nil).RequireStatus(t, 200)
	require.Equal(t, "ok", string(resp.Body))
}

func TestRegisterAndLogin(t *testing.T) {
	truncateAll(t)

	pid, token := registerAndLogin(t, "alice@example.com", "alice")
	require.NotZero(t, pid)
	require.NotEmpty(t, token)

	// Токен работает.
	me := do(t, "GET", "/api/v1/me", token, nil).RequireStatus(t, 200)
	var got struct {
		ID       int64  `json:"id"`
		Nickname string `json:"nickname"`
	}
	me.JSON(t, &got)
	require.Equal(t, pid, got.ID)
	require.Equal(t, "alice", got.Nickname)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	truncateAll(t)

	registerAndLogin(t, "alice@example.com", "alice")

	resp := do(t, "POST", "/api/v1/auth/register", "", map[string]any{
		"full_name": "Other",
		"email":     "alice@example.com",
		"password":  "secret123",
		"nickname":  "other",
		"city":      "Moscow",
	})
	resp.RequireStatus(t, 409)
}

func TestRegister_DuplicateNickname(t *testing.T) {
	truncateAll(t)

	registerAndLogin(t, "alice@example.com", "alice")

	resp := do(t, "POST", "/api/v1/auth/register", "", map[string]any{
		"full_name": "Other",
		"email":     "other@example.com",
		"password":  "secret123",
		"nickname":  "alice",
		"city":      "Moscow",
	})
	resp.RequireStatus(t, 409)
}

func TestRegister_Validation(t *testing.T) {
	truncateAll(t)

	tests := []struct {
		name string
		body map[string]any
	}{
		{
			name: "empty email",
			body: map[string]any{
				"full_name": "X", "email": "", "password": "secret123",
				"nickname": "x", "city": "M",
			},
		},
		{
			name: "short password",
			body: map[string]any{
				"full_name": "X", "email": "x@x.com", "password": "1",
				"nickname": "x", "city": "M",
			},
		},
		{
			name: "empty nickname",
			body: map[string]any{
				"full_name": "X", "email": "x@x.com", "password": "secret123",
				"nickname": "", "city": "M",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			resp := do(t, "POST", "/api/v1/auth/register", "", tt.body)
			require.Equal(t, 400, resp.Status, "body: %s", string(resp.Body))
		})
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	truncateAll(t)

	registerAndLogin(t, "alice@example.com", "alice")

	resp := do(t, "POST", "/api/v1/auth/login", "", map[string]any{
		"email":    "alice@example.com",
		"password": "wrong",
	})
	resp.RequireStatus(t, 401)
}

func TestLogin_UnknownEmail(t *testing.T) {
	truncateAll(t)

	resp := do(t, "POST", "/api/v1/auth/login", "", map[string]any{
		"email":    "nobody@example.com",
		"password": "secret123",
	})
	resp.RequireStatus(t, 401)
}

func TestUnauthorized(t *testing.T) {
	truncateAll(t)

	tests := []struct {
		name   string
		token  string
		status int
	}{
		{"no token", "", 401},
		{"bad token", "not-a-jwt", 401},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			resp := do(t, "GET", "/api/v1/me", tt.token, nil)
			require.Equal(t, tt.status, resp.Status)
		})
	}
}
