package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/division"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

type matchFix struct {
	seasonID   shared.SeasonID
	divisionID shared.DivisionID
	blackID    shared.PlayerID
	whiteID    shared.PlayerID
}

func setupMatchFix(t *testing.T) matchFix {
	t.Helper()
	s := mkSeason(t, 1)
	dr := postgres.NewDivisionRepo(testPool)
	d, _ := division.New(s.ID, "A", "A", 0, 2000, 1, time.Now().UTC())
	require.NoError(t, dr.Create(context.Background(), d))

	pr := postgres.NewPlayerRepo(testPool)
	black := mkPlayer(t, "black", "b@x.com")
	white := mkPlayer(t, "white", "w@x.com")
	require.NoError(t, pr.Create(context.Background(), black))
	require.NoError(t, pr.Create(context.Background(), white))

	return matchFix{seasonID: s.ID, divisionID: d.ID, blackID: black.ID, whiteID: white.ID}
}

func mkMatch(t *testing.T, f matchFix, status dommatch.Status) *dommatch.Match {
	t.Helper()
	m, err := dommatch.New(dommatch.NewMatchParams{
		SeasonID:      f.seasonID,
		DivisionID:    f.divisionID,
		BlackPlayerID: f.blackID,
		WhitePlayerID: f.whiteID,
		CreatedBy:     f.blackID,
		Server:        dommatch.ServerOGS,
	}, time.Now().UTC())
	require.NoError(t, err)
	m.Status = status
	return m
}

func TestMatchRepo_CRUD(t *testing.T) {
	truncateAll(t)
	f := setupMatchFix(t)
	repo := postgres.NewMatchRepo(testPool)

	m := mkMatch(t, f, dommatch.StatusProposed)
	require.NoError(t, repo.Create(context.Background(), m))
	require.NotZero(t, m.ID)

	got, err := repo.GetByID(context.Background(), m.ID)
	require.NoError(t, err)
	require.Equal(t, dommatch.StatusProposed, got.Status)
	require.Nil(t, got.WinnerID)

	// Update: подтверждаем.
	winner := f.blackID
	score := 3.5
	now := time.Now().UTC().Truncate(time.Microsecond)
	m.Status = dommatch.StatusConfirmed
	m.WinnerID = &winner
	m.ScoreDiff = &score
	m.FinishedAt = &now
	m.ConfirmedAt = &now
	require.NoError(t, repo.Update(context.Background(), m))

	got, _ = repo.GetByID(context.Background(), m.ID)
	require.Equal(t, dommatch.StatusConfirmed, got.Status)
	require.NotNil(t, got.WinnerID)
	require.Equal(t, winner, *got.WinnerID)
	require.NotNil(t, got.ScoreDiff)
	require.Equal(t, 3.5, *got.ScoreDiff)
	require.NotNil(t, got.FinishedAt)
}

func TestMatchRepo_CountBetweenPlayersInSeason(t *testing.T) {
	truncateAll(t)
	f := setupMatchFix(t)
	repo := postgres.NewMatchRepo(testPool)

	// 0 подтверждённых.
	n, err := repo.CountBetweenPlayersInSeason(context.Background(), f.seasonID, f.blackID, f.whiteID)
	require.NoError(t, err)
	require.Equal(t, 0, n)

	// 3 confirmed + 1 proposed.
	for i := 0; i < 3; i++ {
		m := mkMatch(t, f, dommatch.StatusConfirmed)
		require.NoError(t, repo.Create(context.Background(), m))
	}
	proposed := mkMatch(t, f, dommatch.StatusProposed)
	require.NoError(t, repo.Create(context.Background(), proposed))

	n, err = repo.CountBetweenPlayersInSeason(context.Background(), f.seasonID, f.blackID, f.whiteID)
	require.NoError(t, err)
	require.Equal(t, 3, n) // proposed не считается
}

func TestMatchRepo_CountConfirmedByPlayerInSeason(t *testing.T) {
	truncateAll(t)
	f := setupMatchFix(t)
	repo := postgres.NewMatchRepo(testPool)

	for i := 0; i < 4; i++ {
		m := mkMatch(t, f, dommatch.StatusConfirmed)
		require.NoError(t, repo.Create(context.Background(), m))
	}

	n, err := repo.CountConfirmedByPlayerInSeason(context.Background(), f.seasonID, f.blackID)
	require.NoError(t, err)
	require.Equal(t, 4, n)

	n, _ = repo.CountConfirmedByPlayerInSeason(context.Background(), f.seasonID, f.whiteID)
	require.Equal(t, 4, n)
}

func TestMatchRepo_ListByPlayerAndDivision(t *testing.T) {
	truncateAll(t)
	f := setupMatchFix(t)
	repo := postgres.NewMatchRepo(testPool)

	m1 := mkMatch(t, f, dommatch.StatusConfirmed)
	m2 := mkMatch(t, f, dommatch.StatusPlaying)
	require.NoError(t, repo.Create(context.Background(), m1))
	require.NoError(t, repo.Create(context.Background(), m2))

	// ListByPlayer: black — видит оба.
	list, err := repo.ListByPlayer(context.Background(), f.blackID, ports.MatchFilter{})
	require.NoError(t, err)
	require.Len(t, list, 2)

	// Фильтр по статусу.
	confirmed := dommatch.StatusConfirmed
	list, _ = repo.ListByPlayer(context.Background(), f.blackID,
		ports.MatchFilter{Status: &confirmed})
	require.Len(t, list, 1)
	require.Equal(t, m1.ID, list[0].ID)

	// ListByDivision: видит оба.
	list, _ = repo.ListByDivision(context.Background(), f.divisionID, ports.MatchFilter{})
	require.Len(t, list, 2)
}
