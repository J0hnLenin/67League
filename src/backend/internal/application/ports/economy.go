package ports

import (
	"context"

	"github.com/J0hnLenin/67League/internal/domain/economy"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// StoneTransactionRepository — журнал транзакций камней.
type StoneTransactionRepository interface {
	Create(ctx context.Context, t *economy.StoneTransaction) error
	ListByPlayer(ctx context.Context, playerID shared.PlayerID, p Pagination) ([]*economy.StoneTransaction, error)
}

// PurchaseRepository — покупки.
type PurchaseRepository interface {
	Create(ctx context.Context, p *economy.Purchase) error
	ListByPlayer(ctx context.Context, playerID shared.PlayerID, p Pagination) ([]*economy.Purchase, error)
}
