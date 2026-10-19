package shared_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

func TestNewEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"успех: простой email", "alice@example.com", "alice@example.com", false},
		{"успех: с точкой", "bob.smith@example.co.uk", "bob.smith@example.co.uk", false},
		{"успех: с цифрами", "user123@test.io", "user123@test.io", false},
		{"успех: uppercase → lowercase", "ALICE@EXAMPLE.COM", "alice@example.com", false},
		{"успех: с пробелами → trim", "  alice@example.com  ", "alice@example.com", false},
		{"ошибка: без @", "aliceexample.com", "", true},
		{"ошибка: без домена", "alice@", "", true},
		{"ошибка: без локальной части", "@example.com", "", true},
		{"ошибка: без точки в домене", "alice@example", "", true},
		{"ошибка: пустая строка", "", "", true},
		{"ошибка: только пробелы", "   ", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			email, err := shared.NewEmail(tt.input)
			if tt.wantErr {
				require.ErrorIs(t, err, shared.ErrValidation)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(email))
		})
	}
}

func TestNewNickname(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"успех: минимум 2 символа", "ab", "ab", false},
		{"успех: 32 символа", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
		{"успех: с пробелами → trim", "  alice  ", "alice", false},
		{"ошибка: 1 символ", "a", "", true},
		{"ошибка: 33 символа", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "", true},
		{"ошибка: пустая строка", "", "", true},
		{"ошибка: только пробелы", "   ", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			nick, err := shared.NewNickname(tt.input)
			if tt.wantErr {
				require.ErrorIs(t, err, shared.ErrValidation)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(nick))
		})
	}
}

func TestNewRating(t *testing.T) {
	t.Parallel()

	t.Run("успех: положительный", func(t *testing.T) {
		t.Parallel()
		r, err := shared.NewRating(1500)
		require.NoError(t, err)
		require.Equal(t, shared.Rating(1500), r)
	})

	t.Run("успех: ноль", func(t *testing.T) {
		t.Parallel()
		r, err := shared.NewRating(0)
		require.NoError(t, err)
		require.Equal(t, shared.Rating(0), r)
		require.True(t, r.IsZero())
	})

	t.Run("ошибка: отрицательный", func(t *testing.T) {
		t.Parallel()
		_, err := shared.NewRating(-1)
		require.ErrorIs(t, err, shared.ErrValidation)
	})
}

func TestPlayerID_String(t *testing.T) {
	t.Parallel()
	require.Equal(t, "42", shared.PlayerID(42).String())
	require.Equal(t, "0", shared.PlayerID(0).String())
}

func TestPlayerID_IsZero(t *testing.T) {
	t.Parallel()
	require.True(t, shared.PlayerID(0).IsZero())
	require.False(t, shared.PlayerID(1).IsZero())
}
