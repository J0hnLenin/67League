package economy

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// StoneTransaction — запись в истории камней.
type StoneTransaction struct {
	ID        shared.StoneTransactionID
	PlayerID  shared.PlayerID
	Amount    shared.Stones
	Reason    string
	RefID     *string // ID партии / бейджа / покупки
	CreatedAt time.Time
}

func NewTransaction(playerID shared.PlayerID, amount shared.Stones, reason string, refID *string, now time.Time) *StoneTransaction {
	return &StoneTransaction{
		ID:        shared.NewStoneTransactionID(),
		PlayerID:  playerID,
		Amount:    amount,
		Reason:    reason,
		RefID:     refID,
		CreatedAt: now,
	}
}

// Purchase — покупка в магазине.
type Purchase struct {
	ID        shared.PurchaseID
	PlayerID  shared.PlayerID
	ItemCode  string
	Price     shared.Stones
	CreatedAt time.Time
}

func NewPurchase(playerID shared.PlayerID, itemCode string, price shared.Stones, now time.Time) (*Purchase, error) {
	if itemCode == "" || price < 0 {
		return nil, shared.ErrValidation
	}
	return &Purchase{
		ID:        shared.NewPurchaseID(),
		PlayerID:  playerID,
		ItemCode:  itemCode,
		Price:     price,
		CreatedAt: now,
	}, nil
}
