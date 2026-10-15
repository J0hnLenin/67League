package ports

import (
	"context"

	"github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// MatchRepository — доступ к партиям.
//
// Важно: методы CountBetweenPlayersInSeason и CountConfirmedByPlayerInSeason
// вызываются внутри транзакции при создании/подтверждении партии, чтобы
// гарантировать соблюдение лимита n партий с одним соперником.
type MatchRepository interface {
	Create(ctx context.Context, m *match.Match) error

	GetByID(ctx context.Context, id shared.MatchID) (*match.Match, error)
	Update(ctx context.Context, m *match.Match) error

	// ListByPlayer — партии игрока с фильтром.
	ListByPlayer(ctx context.Context, playerID shared.PlayerID, f MatchFilter) ([]*match.Match, error)

	// ListByDivision — все партии дивизиона (для подсчёта и админки).
	ListByDivision(ctx context.Context, divisionID shared.DivisionID, f MatchFilter) ([]*match.Match, error)

	// CountBetweenPlayersInSeason — число партий между двумя игроками в сезоне
	// (только confirmed). Используется для проверки лимита n.
	CountBetweenPlayersInSeason(ctx context.Context, seasonID shared.SeasonID, p1, p2 shared.PlayerID) (int, error)

	// CountConfirmedByPlayerInSeason — сколько партий игрок уже подтвердил в сезоне.
	// Используется для расчёта бонуса «сверх t».
	CountConfirmedByPlayerInSeason(ctx context.Context, seasonID shared.SeasonID, playerID shared.PlayerID) (int, error)

	// CountByPlayerAndOpponentPerServer — на будущее (аналитика).
	CountByPlayerAndOpponentPerServer(ctx context.Context, seasonID shared.SeasonID, p1, p2 shared.PlayerID, server match.Server) (int, error)
}

// MatchFilter — фильтр списка партий.
type MatchFilter struct {
	Status *match.Status
	Server *match.Server

	Pagination Pagination
}

// MatchEventRepository — журнал изменений партии.
type MatchEventRepository interface {
	Create(ctx context.Context, e *match.MatchEvent) error
	ListByMatch(ctx context.Context, matchID shared.MatchID) ([]*match.MatchEvent, error)
}
