package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

func TestMatchEventRepo_CreateAndList(t *testing.T) {
	truncateAll(t)
	f := setupMatchFix(t)
	mr := postgres.NewMatchRepo(testPool)
	er := postgres.NewMatchEventRepo(testPool)

	m := mkMatch(t, f, dommatch.StatusProposed)
	require.NoError(t, mr.Create(context.Background(), m))

	now := time.Now().UTC().Truncate(time.Microsecond)

	e1 := dommatch.NewEvent(m.ID, dommatch.EventCreated, f.blackID, nil, now)
	e2 := dommatch.NewEvent(m.ID, dommatch.EventAccepted, f.whiteID,
		map[string]any{"note": "ok"}, now.Add(time.Second))

	require.NoError(t, er.Create(context.Background(), e1))
	require.NoError(t, er.Create(context.Background(), e2))
	require.NotZero(t, e1.ID)

	list, err := er.ListByMatch(context.Background(), m.ID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, dommatch.EventCreated, list[0].Type)
	require.Equal(t, dommatch.EventAccepted, list[1].Type)
	require.Equal(t, "ok", list[1].Payload["note"])
}
