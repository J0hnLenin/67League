package season

// Settings — все настраиваемые константы сезона.
// Хранятся в БД в JSONB-поле Season.settings.
type Settings struct {
	// Ротация
	K1 int `json:"k1"` // S ↔ A
	K2 int `json:"k2"` // A ↔ B
	K3 int `json:"k3"` // B ↔ C (на будущее)

	// Зачёт
	P int `json:"p"` // минимум партий для зачёта
	N int `json:"n"` // максимум партий с одним соперником
	T int `json:"t"` // порог партий для бонуса +0.5 камня

	// Сроки
	ChallengeAcceptDays int `json:"challenge_accept_days"`
	MatchCancelDays     int `json:"match_cancel_days"`
	SeasonDurationDays  int `json:"season_duration_days"`
	ClosingDays         int `json:"closing_days"`

	// LP
	LPWin  int `json:"lp_win"`
	LPLoss int `json:"lp_loss"`

	// Бонусы за снятие стрика (в LP и камнях одновременно)
	StreakBonus3  int `json:"streak_bonus_3"`
	StreakBonus5  int `json:"streak_bonus_5"`
	StreakBonus67 int `json:"streak_bonus_67"` // 6 или 7

	// Камни
	StonesWin           float64 `json:"stones_win"`
	StonesLoss          float64 `json:"stones_loss"`
	StonesClubTop1      float64 `json:"stones_club_top1"`
	StonesSuperTitle    float64 `json:"stones_super_title"`
	StonesPerExtraMatch float64 `json:"stones_per_extra_match"` // +0.5 по умолчанию
}

// DefaultSettings — значения по умолчанию для нового сезона.
func DefaultSettings() Settings {
	return Settings{
		K1: 1, K2: 2, K3: 2,
		P: 8, N: 4, T: 12,
		ChallengeAcceptDays: 3,
		MatchCancelDays:     3,
		SeasonDurationDays:  30,
		ClosingDays:         3,
		LPWin:               3,
		LPLoss:              0,
		StreakBonus3:        3,
		StreakBonus5:        5,
		StreakBonus67:       6, // или 7 — решается случайно в момент начисления
		StonesWin:           3,
		StonesLoss:          0,
		StonesClubTop1:      5,
		StonesSuperTitle:    67,
		StonesPerExtraMatch: 0.5,
	}
}

// Validate проверяет корректность настроек.
func (s Settings) Validate() error {
	if s.P < 0 || s.N <= 0 || s.T < 0 {
		return ErrInvalidSettings
	}
	if s.K1 < 0 || s.K2 < 0 || s.K3 < 0 {
		return ErrInvalidSettings
	}
	if s.StreakBonus67 != 6 && s.StreakBonus67 != 7 {
		return ErrInvalidSettings
	}
	return nil
}
