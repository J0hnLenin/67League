package division

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type MemberStatus string

const (
	MemberActive   MemberStatus = "active"
	MemberExcluded MemberStatus = "excluded"
)

// DivisionMember — участие игрока в дивизионе конкретного сезона.
type DivisionMember struct {
	ID         shared.DivisionMemberID
	DivisionID shared.DivisionID
	PlayerID   shared.PlayerID

	LP            shared.LP
	Wins          int
	Losses        int
	MatchesPlayed int
	StreakCurrent int
	StreakMax     int
	FinalPlace    *int

	Status    MemberStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewMember(divisionID shared.DivisionID, playerID shared.PlayerID, now time.Time) *DivisionMember {
	return &DivisionMember{
		ID:         shared.NewDivisionMemberID(),
		DivisionID: divisionID,
		PlayerID:   playerID,
		Status:     MemberActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// RegisterWin — победа. Возвращает бонус за снятие стрика у соперника.
// Значение opponentStreakBeforeWin — стрик соперника до этой партии.
func (m *DivisionMember) RegisterWin(settings StreakSettings, opponentStreakBeforeWin int, now time.Time) (lpBonus, stoneBonus shared.LP) {
	m.Wins++
	m.MatchesPlayed++
	m.StreakCurrent++
	if m.StreakCurrent > m.StreakMax {
		m.StreakMax = m.StreakCurrent
	}

	switch {
	case opponentStreakBeforeWin >= 6:
		return shared.LP(settings.StreakBonus67), shared.LP(settings.StreakBonus67)
	case opponentStreakBeforeWin == 5:
		return shared.LP(settings.StreakBonus5), shared.LP(settings.StreakBonus5)
	case opponentStreakBeforeWin == 3 || opponentStreakBeforeWin == 4:
		return shared.LP(settings.StreakBonus3), shared.LP(settings.StreakBonus3)
	}
	return 0, 0
}

// RegisterLoss — поражение обнуляет текущий стрик.
func (m *DivisionMember) RegisterLoss(now time.Time) {
	m.Losses++
	m.MatchesPlayed++
	m.StreakCurrent = 0
	m.UpdatedAt = now
}

func (m *DivisionMember) AddLP(lp shared.LP, now time.Time) {
	m.LP += lp
	m.UpdatedAt = now
}

func (m *DivisionMember) Exclude(now time.Time) {
	m.Status = MemberExcluded
	m.UpdatedAt = now
}

// StreakSettings — минимальный интерфейс, чтобы не тянуть весь season.Settings.
type StreakSettings struct {
	StreakBonus3  int
	StreakBonus5  int
	StreakBonus67 int
}
