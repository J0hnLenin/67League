package player_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

func TestNewPlayer(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		params  player.NewPlayerParams
		wantErr bool
	}{
		{
			name: "успех: минимальные поля",
			params: player.NewPlayerParams{
				FullName:     "Alice",
				Email:        "alice@example.com",
				PasswordHash: "hash",
				Nickname:     "alice",
				City:         "Moscow",
			},
			wantErr: false,
		},
		{
			name: "успех: со всеми полями",
			params: player.NewPlayerParams{
				FullName:     "Bob",
				Email:        "bob@example.com",
				PasswordHash: "hash",
				Nickname:     "bob",
				City:         "SPb",
				Rating:       1500,
				KGSNick:      "bob_kgs",
				OGSNick:      "bob_ogs",
				FoxNick:      "bob_fox",
				RFGID:        "42",
				SputnikID:    "99",
			},
			wantErr: false,
		},
		{
			name: "ошибка: пустое имя",
			params: player.NewPlayerParams{
				FullName:     "",
				Email:        "alice@example.com",
				PasswordHash: "hash",
				Nickname:     "alice",
				City:         "Moscow",
			},
			wantErr: true,
		},
		{
			name: "ошибка: пустой город",
			params: player.NewPlayerParams{
				FullName:     "Alice",
				Email:        "alice@example.com",
				PasswordHash: "hash",
				Nickname:     "alice",
				City:         "",
			},
			wantErr: true,
		},
		{
			name: "ошибка: пустой email",
			params: player.NewPlayerParams{
				FullName:     "Alice",
				Email:        "",
				PasswordHash: "hash",
				Nickname:     "alice",
				City:         "Moscow",
			},
			wantErr: true,
		},
		{
			name: "ошибка: пустой ник",
			params: player.NewPlayerParams{
				FullName:     "Alice",
				Email:        "alice@example.com",
				PasswordHash: "hash",
				Nickname:     "",
				City:         "Moscow",
			},
			wantErr: true,
		},
		{
			name: "ошибка: пустой хеш пароля",
			params: player.NewPlayerParams{
				FullName:     "Alice",
				Email:        "alice@example.com",
				PasswordHash: "",
				Nickname:     "alice",
				City:         "Moscow",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := player.New(tt.params, now)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, p)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, p)
			require.Equal(t, tt.params.FullName, p.FullName)
			require.Equal(t, tt.params.Email, p.Email)
			require.Equal(t, tt.params.Nickname, p.Nickname)
			require.Equal(t, tt.params.City, p.City)
			require.Equal(t, tt.params.Rating, p.Rating)
			require.Equal(t, player.StatusActive, p.Status)
			require.True(t, p.CreatedAt.Equal(now))
			require.True(t, p.UpdatedAt.Equal(now))
			require.Zero(t, p.ID)
			require.Nil(t, p.ClubID)
			require.Zero(t, p.Stones)
			require.Zero(t, p.LPTotal)
		})
	}
}

func TestPlayer_Ban(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: бан активного игрока", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.Ban(now.Add(time.Hour))
		require.NoError(t, err)
		require.Equal(t, player.StatusBanned, p.Status)
	})

	t.Run("ошибка: повторный бан", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.Ban(now.Add(time.Hour)))
		err := p.Ban(now.Add(2 * time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestPlayer_Unban(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: разбан", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.Ban(now.Add(time.Hour)))
		err := p.Unban(now.Add(2 * time.Hour))
		require.NoError(t, err)
		require.Equal(t, player.StatusActive, p.Status)
	})

	t.Run("ошибка: разбан активного", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.Unban(now)
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestPlayer_MarkInactive(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: маркировка неактивным", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.MarkInactive(now.Add(time.Hour))
		require.NoError(t, err)
		require.Equal(t, player.StatusInactive, p.Status)
	})

	t.Run("ошибка: неактивный уже неактивный", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.MarkInactive(now.Add(time.Hour)))
		err := p.MarkInactive(now.Add(2 * time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestPlayer_Reactivate(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: реактивация", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.MarkInactive(now.Add(time.Hour)))
		err := p.Reactivate(now.Add(2 * time.Hour))
		require.NoError(t, err)
		require.Equal(t, player.StatusActive, p.Status)
	})

	t.Run("ошибка: реактивация активного", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.Reactivate(now)
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestPlayer_JoinClub(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	clubID := shared.ClubID(42)
	otherClubID := shared.ClubID(43)

	t.Run("успех: вступление в клуб", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.JoinClub(clubID, now.Add(time.Hour))
		require.NoError(t, err)
		require.NotNil(t, p.ClubID)
		require.Equal(t, clubID, *p.ClubID)
	})

	t.Run("ошибка: уже в этом клубе", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.JoinClub(clubID, now.Add(time.Hour)))
		err := p.JoinClub(clubID, now.Add(2*time.Hour))
		require.ErrorIs(t, err, shared.ErrConflict)
	})

	t.Run("успех: переход в другой клуб", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.JoinClub(clubID, now.Add(time.Hour)))
		err := p.JoinClub(otherClubID, now.Add(2*time.Hour))
		require.NoError(t, err)
		require.Equal(t, otherClubID, *p.ClubID)
	})

	t.Run("ошибка: забаненный не может вступить", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.Ban(now.Add(time.Hour)))
		err := p.JoinClub(clubID, now.Add(2*time.Hour))
		require.ErrorIs(t, err, shared.ErrForbidden)
	})
}

func TestPlayer_LeaveClub(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	clubID := shared.ClubID(42)

	t.Run("успех: выход из клуба", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.JoinClub(clubID, now.Add(time.Hour)))
		err := p.LeaveClub(now.Add(2 * time.Hour))
		require.NoError(t, err)
		require.Nil(t, p.ClubID)
	})

	t.Run("ошибка: не в клубе", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.LeaveClub(now)
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestPlayer_SetTitle(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	titleID := shared.TitleID(7)

	t.Run("установка титула", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		p.SetTitle(&titleID, now.Add(time.Hour))
		require.NotNil(t, p.ActiveTitleID)
		require.Equal(t, titleID, *p.ActiveTitleID)
	})

	t.Run("снятие титула", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		p.SetTitle(&titleID, now.Add(time.Hour))
		p.SetTitle(nil, now.Add(2*time.Hour))
		require.Nil(t, p.ActiveTitleID)
	})
}

func TestPlayer_AddStones(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: начисление", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.AddStones(10, now)
		require.NoError(t, err)
		require.Equal(t, shared.Stones(10), p.Stones)
	})

	t.Run("успех: списание", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.AddStones(10, now))
		err := p.AddStones(-5, now)
		require.NoError(t, err)
		require.Equal(t, shared.Stones(5), p.Stones)
	})

	t.Run("успех: ноль", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.AddStones(0, now)
		require.NoError(t, err)
		require.Zero(t, p.Stones)
	})

	t.Run("ошибка: недостаточно камней", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.AddStones(-5, now)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient stones")
	})
}

func TestPlayer_AddLP(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("начисление LP", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		p.AddLP(3, now)
		require.Equal(t, shared.LP(3), p.LPTotal)
	})

	t.Run("накопление LP", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		p.AddLP(3, now)
		p.AddLP(5, now)
		require.Equal(t, shared.LP(8), p.LPTotal)
	})
}

func TestPlayer_CanCreateMatchWith(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	otherID := shared.PlayerID(999)

	t.Run("успех: активный игрок", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		err := p.CanCreateMatchWith(otherID)
		require.NoError(t, err)
	})

	t.Run("ошибка: сам с собой", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		// ID ещё 0, поэтому используем другой подход
		p.ID = 42
		err := p.CanCreateMatchWith(p.ID)
		require.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("ошибка: забаненный", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.Ban(now.Add(time.Hour)))
		err := p.CanCreateMatchWith(otherID)
		require.ErrorIs(t, err, shared.ErrForbidden)
	})

	t.Run("ошибка: неактивный", func(t *testing.T) {
		t.Parallel()
		p := mustNewPlayer(t, now)
		require.NoError(t, p.MarkInactive(now.Add(time.Hour)))
		err := p.CanCreateMatchWith(otherID)
		require.ErrorIs(t, err, shared.ErrForbidden)
	})
}

func TestStatus_CanPlay(t *testing.T) {
	t.Parallel()
	require.True(t, player.StatusActive.CanPlay())
	require.False(t, player.StatusInactive.CanPlay())
	require.False(t, player.StatusBanned.CanPlay())
}

func TestStatus_CanCreateClub(t *testing.T) {
	t.Parallel()
	require.True(t, player.StatusActive.CanCreateClub())
	require.False(t, player.StatusInactive.CanCreateClub())
	require.False(t, player.StatusBanned.CanCreateClub())
}

func TestStatus_CanJoinSeason(t *testing.T) {
	t.Parallel()
	require.True(t, player.StatusActive.CanJoinSeason())
	require.False(t, player.StatusInactive.CanJoinSeason())
	require.False(t, player.StatusBanned.CanJoinSeason())
}

func TestStatus_Valid(t *testing.T) {
	t.Parallel()
	require.True(t, player.StatusActive.Valid())
	require.True(t, player.StatusInactive.Valid())
	require.True(t, player.StatusBanned.Valid())
	require.False(t, player.Status("unknown").Valid())
}

// ---- helpers ----------------------------------------------------------------

func mustNewPlayer(t *testing.T, now time.Time) *player.Player {
	t.Helper()
	p, err := player.New(player.NewPlayerParams{
		FullName:     "Test Player",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Nickname:     "test",
		City:         "Moscow",
	}, now)
	require.NoError(t, err)
	return p
}
