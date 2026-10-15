package dto

import (
	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// RegisterPlayerInput — вход регистрации. Все значения уже типизированы.
// Строки → типы конвертирует delivery-слой.
type RegisterPlayerInput struct {
	FullName  string
	Email     shared.Email
	Password  string
	Nickname  shared.Nickname
	City      string
	KGSNick   string
	OGSNick   string
	FoxNick   string
	RFGID     string
	SputnikID string
	ClubID    *shared.ClubID
}

type RegisterPlayerOutput struct {
	Player   *player.Player
	Division *shared.DivisionID // nil, если игрок не попал в активный сезон
}

type LoginPlayerInput struct {
	Email    shared.Email
	Password string
}

type LoginPlayerOutput struct {
	Player *player.Player
	Token  string
}
