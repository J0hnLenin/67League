package badge

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type Type string

const (
	TypePersonal Type = "personal"
	TypeSeasonal Type = "seasonal"
	TypeClub     Type = "club"
	TypeSuper    Type = "super"
)

func (t Type) Valid() bool {
	switch t {
	case TypePersonal, TypeSeasonal, TypeClub, TypeSuper:
		return true
	}
	return false
}

type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityRare      Rarity = "rare"
	RarityEpic      Rarity = "epic"
	RarityLegendary Rarity = "legendary"
	Rarity67        Rarity = "6-7"
)

// Badge — справочник бейджей.
type Badge struct {
	ID          shared.BadgeID
	Code        string
	Icon        string
	Name        string
	Description string
	Type        Type
	Condition   map[string]any
	Rarity      Rarity
	TitleID     *shared.TitleID
	CreatedAt   time.Time
}

// Title — справочник титулов. Каждый бейдж может открывать титул.
type Title struct {
	ID      shared.TitleID
	Code    string
	Name    string
	BadgeID *shared.BadgeID
}
