package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/admin"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

func TestAdminRepo_CRUD(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewAdminRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	a, err := admin.New("root", "hashed:secret", now)
	require.NoError(t, err)
	require.NoError(t, repo.Create(context.Background(), a))
	require.NotZero(t, a.ID)

	got, err := repo.GetByID(context.Background(), a.ID)
	require.NoError(t, err)
	require.Equal(t, "root", got.Login)

	gotByLogin, err := repo.GetByLogin(context.Background(), "root")
	require.NoError(t, err)
	require.Equal(t, a.ID, gotByLogin.ID)

	_, err = repo.GetByLogin(context.Background(), "nope")
	require.True(t, errors.Is(err, shared.ErrNotFound))
}

func TestAdminRepo_DuplicateLogin(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewAdminRepo(testPool)
	now := time.Now().UTC()

	a1, _ := admin.New("root", "h1", now)
	a2, _ := admin.New("root", "h2", now)
	require.NoError(t, repo.Create(context.Background(), a1))

	err := repo.Create(context.Background(), a2)
	require.True(t, errors.Is(err, shared.ErrAlreadyExists))
}
