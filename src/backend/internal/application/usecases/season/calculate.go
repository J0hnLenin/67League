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
	"github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// CalculateSeasonResults подводит итоги сезона.
//
// Побочные эффекты (одна транзакция):
//   - каждому DivisionMember проставляется FinalPlace по LP;
//   - Member.Status = excluded для тех, кто сыграл < p партий;
//   - возвращается клубный зачёт и лучший клуб сезона.
//
// Идемпотентно: повторный вызов перезапишет FinalPlace теми же значениями.
type CalculateSeasonResults struct {
	seasons   ports.SeasonRepository
	divisions ports.DivisionRepository
	members   ports.DivisionMemberRepository
	players   ports.PlayerRepository
	clubs     ports.ClubRepository
	clock     ports.Clock
	tx        ports.TransactionManager
}

type CalculateSeasonResultsDeps struct {
	Seasons   ports.SeasonRepository
	Divisions ports.DivisionRepository
	Members   ports.DivisionMemberRepository
	Players   ports.PlayerRepository
	Clubs     ports.ClubRepository
	Clock     ports.Clock
	Tx        ports.TransactionManager
}

func NewCalculateSeasonResults(d CalculateSeasonResultsDeps) (*CalculateSeasonResults, error) {
	if d.Seasons == nil || d.Divisions == nil || d.Members == nil ||
		d.Players == nil || d.Clubs == nil || d.Clock == nil || d.Tx == nil {
		return nil, errors.New("calculate season results: missing dependencies")
	}
	return &CalculateSeasonResults{
		seasons:   d.Seasons,
		divisions: d.Divisions,
		members:   d.Members,
		players:   d.Players,
		clubs:     d.Clubs,
		clock:     d.Clock,
		tx:        d.Tx,
	}, nil
}

func (uc *CalculateSeasonResults) Execute(
	ctx context.Context,
	in dto.CalculateSeasonResultsInput,
) (*dto.CalculateSeasonResultsOutput, error) {
	s, err := uc.seasons.GetByID(ctx, in.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("calculate season results: get season: %w", err)
	}
	if s.Status != season.StatusFinished {
		return nil, fmt.Errorf(
			"calculate season results: season status %q: %w",
			s.Status, shared.ErrInvalidState,
		)
	}

	divs, err := uc.divisions.ListBySeason(ctx, in.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("calculate season results: list divisions: %w", err)
	}
	if len(divs) == 0 {
		return nil, fmt.Errorf("calculate season results: no divisions: %w", shared.ErrInvalidState)
	}

	now := uc.clock.Now()
	out := &dto.CalculateSeasonResultsOutput{Season: s}

	clubAcc := map[shared.ClubID]*clubAccumulator{}

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		for _, d := range divs {
			rows, err := uc.processDivision(ctx, d, s.Settings.P, clubAcc, now)
			if err != nil {
				return err
			}
			out.Standings = append(out.Standings, dto.DivisionStanding{
				Division: d,
				Rows:     rows,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	clubTable, best, err := uc.buildClubTable(ctx, clubAcc)
	if err != nil {
		return nil, err
	}
	out.ClubTable = clubTable
	out.BestClub = best

	return out, nil
}

// ---- Внутренние помощники --------------------------------------------------

type clubAccumulator struct {
	clubID  shared.ClubID
	total   shared.LP
	lpList  []shared.LP
	members int
}

func (uc *CalculateSeasonResults) processDivision(
	ctx context.Context,
	d *division.Division,
	minMatches int,
	acc map[shared.ClubID]*clubAccumulator,
	now time.Time,
) ([]dto.MemberStanding, error) {
	ms, err := uc.members.ListByDivision(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("list members of %s: %w", d.Code, err)
	}

	sortMembersByRank(ms)

	rows := make([]dto.MemberStanding, 0, len(ms))
	for i, m := range ms {
		if m.MatchesPlayed < minMatches {
			m.Exclude(now)
			if err := uc.members.Update(ctx, m); err != nil {
				return nil, fmt.Errorf("exclude member: %w", err)
			}
			rows = append(rows, dto.MemberStanding{
				Member: m, PlayerID: m.PlayerID, Excluded: true,
			})
			continue
		}

		place := i + 1
		m.FinalPlace = &place
		if err := uc.members.Update(ctx, m); err != nil {
			return nil, fmt.Errorf("save member place: %w", err)
		}
		rows = append(rows, dto.MemberStanding{
			Member: m, PlayerID: m.PlayerID, Place: place,
		})

		if err := uc.addToClubAccumulator(ctx, m, acc); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func (uc *CalculateSeasonResults) addToClubAccumulator(
	ctx context.Context,
	m *division.DivisionMember,
	acc map[shared.ClubID]*clubAccumulator,
) error {
	p, err := uc.players.GetByID(ctx, m.PlayerID)
	if err != nil {
		return fmt.Errorf("get player %d: %w", m.PlayerID, err)
	}
	if p.ClubID == nil {
		return nil
	}
	a := acc[*p.ClubID]
	if a == nil {
		a = &clubAccumulator{clubID: *p.ClubID}
		acc[*p.ClubID] = a
	}
	a.total += m.LP
	a.lpList = append(a.lpList, m.LP)
	a.members++
	return nil
}

func (uc *CalculateSeasonResults) buildClubTable(
	ctx context.Context,
	acc map[shared.ClubID]*clubAccumulator,
) ([]dto.ClubStanding, *dto.ClubStanding, error) {
	if len(acc) == 0 {
		return nil, nil, nil
	}

	out := make([]dto.ClubStanding, 0, len(acc))
	for _, a := range acc {
		c, err := uc.clubs.GetByID(ctx, a.clubID)
		if err != nil {
			return nil, nil, fmt.Errorf("get club %d: %w", a.clubID, err)
		}
		out = append(out, dto.ClubStanding{
			ClubID:  a.clubID,
			Club:    c,
			TotalLP: a.total,
			Members: a.members,
			Top5Sum: top5Sum(a.lpList),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].TotalLP != out[j].TotalLP {
			return out[i].TotalLP > out[j].TotalLP
		}
		return out[i].Top5Sum > out[j].Top5Sum
	})

	best := out[0]
	return out, &best, nil
}

// sortMembersByRank — правило из ТЗ:
//  1. больше LP — выше;
//  2. при равенстве — больше партий — выше;
//  3. при равенстве — больше побед — выше.
func sortMembersByRank(ms []*division.DivisionMember) {
	sort.SliceStable(ms, func(i, j int) bool {
		if ms[i].LP != ms[j].LP {
			return ms[i].LP > ms[j].LP
		}
		if ms[i].MatchesPlayed != ms[j].MatchesPlayed {
			return ms[i].MatchesPlayed > ms[j].MatchesPlayed
		}
		return ms[i].Wins > ms[j].Wins
	})
}

func top5Sum(lps []shared.LP) shared.LP {
	sorted := make([]shared.LP, len(lps))
	copy(sorted, lps)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] > sorted[j] })
	var sum shared.LP
	for i := 0; i < 5 && i < len(sorted); i++ {
		sum += sorted[i]
	}
	return sum
}
