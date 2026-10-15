package season

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/division"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// ApplyRotation переносит игроков из завершённого сезона в новый.
//
// Требования:
//   - prev.Status == finished;
//   - next.Status ∈ {announced, registration};
//   - у обоих сезонов есть дивизионы.
//
// Правила:
//   - участники каждого дивизиона сортируются по LP (тайбрейк: партии, победы);
//   - топ-k повышаются, анти-топ-k понижаются, остальные остаются;
//   - k зависит от пары дивизионов: (0,1)→k1, (1,2)→k2, (2,3)→k3;
//   - верхний дивизион — без повышения, нижний — без понижения;
//   - исключённые (Status=excluded) НЕ переносятся — им нужна
//     дополнительная регистрация через JoinSeason.
//
// Идемпотентен: если DivisionMember уже существует в новом сезоне,
// игрок пропускается без создания дубликата.
type ApplyRotation struct {
	seasons   ports.SeasonRepository
	divisions ports.DivisionRepository
	members   ports.DivisionMemberRepository
	clock     ports.Clock
	tx        ports.TransactionManager
}

type ApplyRotationDeps struct {
	Seasons   ports.SeasonRepository
	Divisions ports.DivisionRepository
	Members   ports.DivisionMemberRepository
	Clock     ports.Clock
	Tx        ports.TransactionManager
}

func NewApplyRotation(d ApplyRotationDeps) (*ApplyRotation, error) {
	if d.Seasons == nil || d.Divisions == nil || d.Members == nil ||
		d.Clock == nil || d.Tx == nil {
		return nil, errors.New("apply rotation: missing dependencies")
	}
	return &ApplyRotation{
		seasons:   d.Seasons,
		divisions: d.Divisions,
		members:   d.Members,
		clock:     d.Clock,
		tx:        d.Tx,
	}, nil
}

func (uc *ApplyRotation) Execute(
	ctx context.Context,
	in dto.ApplyRotationInput,
) (*dto.ApplyRotationOutput, error) {
	if in.PreviousSeasonID == in.NextSeasonID {
		return nil, fmt.Errorf("apply rotation: same season: %w", shared.ErrValidation)
	}

	prev, err := uc.seasons.GetByID(ctx, in.PreviousSeasonID)
	if err != nil {
		return nil, fmt.Errorf("apply rotation: get previous season: %w", err)
	}
	if prev.Status != domseason.StatusFinished {
		return nil, fmt.Errorf(
			"apply rotation: previous season status %q: %w",
			prev.Status, shared.ErrInvalidState,
		)
	}

	next, err := uc.seasons.GetByID(ctx, in.NextSeasonID)
	if err != nil {
		return nil, fmt.Errorf("apply rotation: get next season: %w", err)
	}
	if next.Status != domseason.StatusAnnounced && next.Status != domseason.StatusRegistration {
		return nil, fmt.Errorf(
			"apply rotation: next season status %q: %w",
			next.Status, shared.ErrInvalidState,
		)
	}

	prevDivs, err := uc.divisions.ListBySeason(ctx, prev.ID)
	if err != nil {
		return nil, fmt.Errorf("apply rotation: list prev divisions: %w", err)
	}
	nextDivs, err := uc.divisions.ListBySeason(ctx, next.ID)
	if err != nil {
		return nil, fmt.Errorf("apply rotation: list next divisions: %w", err)
	}
	if len(prevDivs) == 0 || len(nextDivs) == 0 {
		return nil, fmt.Errorf("apply rotation: empty divisions: %w", shared.ErrInvalidState)
	}

	sort.SliceStable(prevDivs, func(i, j int) bool {
		return prevDivs[i].Order < prevDivs[j].Order
	})
	sort.SliceStable(nextDivs, func(i, j int) bool {
		return nextDivs[i].Order < nextDivs[j].Order
	})

	nextByOrder := make(map[int]*division.Division, len(nextDivs))
	for _, d := range nextDivs {
		nextByOrder[d.Order] = d
	}

	// Все целевые позиции должны существовать в новом сезоне.
	for i := range prevDivs {
		if nextByOrder[i] == nil {
			return nil, fmt.Errorf(
				"apply rotation: no division with order %d in next season: %w",
				i, shared.ErrInvalidState,
			)
		}
	}

	now := uc.clock.Now()
	out := &dto.ApplyRotationOutput{}

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		for i, pd := range prevDivs {
			if err := uc.processDivision(
				ctx, pd, i, len(prevDivs),
				prev.Settings, next.ID, nextByOrder, out, now,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// processDivision обрабатывает один дивизион предыдущего сезона.
//
//	idx    — позиция дивизиона в отсортированном списке (0 = верхний).
//	total  — всего дивизионов в предыдущем сезоне.
func (uc *ApplyRotation) processDivision(
	ctx context.Context,
	pd *division.Division,
	idx int,
	total int,
	settings domseason.Settings,
	nextSeasonID shared.SeasonID,
	nextByOrder map[int]*division.Division,
	out *dto.ApplyRotationOutput,
	now time.Time,
) error {
	all, err := uc.members.ListByDivision(ctx, pd.ID)
	if err != nil {
		return fmt.Errorf("apply rotation: list members of %q: %w", pd.Code, err)
	}
	sortMembersByRank(all)

	// Исключённые не переносятся — они должны явно зарегистрироваться
	// в новом сезоне через JoinSeason.
	valid := make([]*division.DivisionMember, 0, len(all))
	for _, m := range all {
		if m.Status == division.MemberExcluded {
			continue
		}
		valid = append(valid, m)
	}
	if len(valid) == 0 {
		return nil
	}

	k := rotationK(settings, idx)

	upCut := k
	if idx == 0 {
		upCut = 0 // верхний дивизион — без повышения
	}
	downCut := k
	if idx == total-1 {
		downCut = 0 // нижний дивизион — без понижения
	}

	// Защита от пересечения up/down при малом числе участников.
	if upCut > len(valid) {
		upCut = len(valid)
	}
	if downCut > len(valid) {
		downCut = len(valid)
	}
	if upCut+downCut > len(valid) {
		// Приоритет — повышению.
		downCut = len(valid) - upCut
	}

	for i, m := range valid {
		place := i + 1

		var (
			dir         dto.RotationDirection
			targetOrder int
		)
		switch {
		case i < upCut:
			dir = dto.RotationUp
			targetOrder = idx - 1
		case i >= len(valid)-downCut:
			dir = dto.RotationDown
			targetOrder = idx + 1
		default:
			dir = dto.RotationStay
			targetOrder = idx
		}

		target := nextByOrder[targetOrder]
		if target == nil {
			return fmt.Errorf(
				"apply rotation: no division with order %d: %w",
				targetOrder, shared.ErrInvalidState,
			)
		}

		created, err := uc.ensureMember(ctx, target.ID, m.PlayerID, nextSeasonID, now)
		if err != nil {
			return err
		}
		if !created {
			continue
		}

		out.Created++
		out.Transfers = append(out.Transfers, dto.PlayerTransfer{
			PlayerID:  m.PlayerID,
			FromDivID: pd.ID,
			ToDivID:   target.ID,
			Direction: dir,
			FromPlace: place,
		})
	}

	return nil
}

// ensureMember создаёт DivisionMember в новом сезоне, если её ещё нет.
// Возвращает true, если запись была создана.
func (uc *ApplyRotation) ensureMember(
	ctx context.Context,
	divID shared.DivisionID,
	playerID shared.PlayerID,
	seasonID shared.SeasonID,
	now time.Time,
) (bool, error) {
	_, err := uc.members.GetByPlayerAndSeason(ctx, playerID, seasonID)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, shared.ErrNotFound) {
		return false, fmt.Errorf("apply rotation: check membership: %w", err)
	}

	m := division.NewMember(divID, playerID, now)
	if err := uc.members.Create(ctx, m); err != nil {
		return false, fmt.Errorf("apply rotation: create member: %w", err)
	}
	return true, nil
}

// rotationK выбирает k1/k2/k3 в зависимости от позиции дивизиона.
//
//	idx = 0 → пара (0,1) → k1
//	idx = 1 → пара (1,2) → k2
//	idx = 2 → пара (2,3) → k3
//
// Для idx ≥ 3 возвращается 0: расширение на большее число дивизионов
// потребует добавить k4, k5, ... в Settings.
func rotationK(s domseason.Settings, idx int) int {
	switch idx {
	case 0:
		return s.K1
	case 1:
		return s.K2
	case 2:
		return s.K3
	default:
		return 0
	}
}
