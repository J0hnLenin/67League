package match

import (
	"context"
	"errors"
	"fmt"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/economy"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/notification"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// ConfirmResult подтверждает результат партии и начисляет LP/камни.
//
// Всё это — одна транзакция:
//  1. Перевести Match в confirmed.
//  2. Обновить статистику обоих игроков в DivisionMember.
//  3. Начислить LP и камни победителю и проигравшему.
//  4. Списать/обновить стрик.
//  5. Записать MatchEvent.
//  6. Создать уведомление соперника.
type ConfirmResult struct {
	matches ports.MatchRepository
	events  ports.MatchEventRepository
	members ports.DivisionMemberRepository
	players ports.PlayerRepository
	seasons ports.SeasonRepository
	stones  ports.StoneTransactionRepository
	notifs  ports.NotificationRepository
	clock   ports.Clock
	tx      ports.TransactionManager
}

type ConfirmResultDeps struct {
	Matches ports.MatchRepository
	Events  ports.MatchEventRepository
	Members ports.DivisionMemberRepository
	Players ports.PlayerRepository
	Seasons ports.SeasonRepository
	Stones  ports.StoneTransactionRepository
	Notifs  ports.NotificationRepository
	Clock   ports.Clock
	Tx      ports.TransactionManager
}

func NewConfirmResult(d ConfirmResultDeps) (*ConfirmResult, error) {
	if d.Matches == nil || d.Events == nil || d.Members == nil ||
		d.Players == nil || d.Seasons == nil || d.Stones == nil ||
		d.Notifs == nil || d.Clock == nil || d.Tx == nil {
		return nil, errors.New("confirm result: missing dependencies")
	}
	return &ConfirmResult{
		matches: d.Matches, events: d.Events, members: d.Members,
		players: d.Players, seasons: d.Seasons, stones: d.Stones,
		notifs: d.Notifs, clock: d.Clock, tx: d.Tx,
	}, nil
}

func (uc *ConfirmResult) Execute(ctx context.Context, in dto.ConfirmResultInput) (*dto.ConfirmResultOutput, error) {
	now := uc.clock.Now()

	var out *dto.ConfirmResultOutput

	err := uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		m, err := uc.matches.GetByID(ctx, in.MatchID)
		if err != nil {
			return fmt.Errorf("confirm result: get match: %w", err)
		}

		if err := m.Confirm(in.ActorID, now); err != nil {
			return fmt.Errorf("confirm result: %w", err)
		}
		if m.WinnerID == nil {
			return fmt.Errorf("confirm result: no winner: %w", shared.ErrInvalidState)
		}
		winnerID := *m.WinnerID
		loserID, ok := m.OpponentOf(winnerID)
		if !ok {
			return fmt.Errorf("confirm result: resolve loser: %w", shared.ErrInvalidState)
		}

		s, err := uc.seasons.GetByID(ctx, m.SeasonID)
		if err != nil {
			return fmt.Errorf("confirm result: get season: %w", err)
		}
		settings := s.Settings

		winnerMember, err := uc.members.GetByPlayerAndSeason(ctx, winnerID, m.SeasonID)
		if err != nil {
			return fmt.Errorf("confirm result: winner member: %w", err)
		}
		loserMember, err := uc.members.GetByPlayerAndSeason(ctx, loserID, m.SeasonID)
		if err != nil {
			return fmt.Errorf("confirm result: loser member: %w", err)
		}

		// Захватываем стрик проигравшего ДО обновления.
		streakBefore := loserMember.StreakCurrent

		// Обновляем статистику.
		lpBonus, stoneBonus := winnerMember.RegisterWin(division.StreakSettings{
			StreakBonus3:  settings.StreakBonus3,
			StreakBonus5:  settings.StreakBonus5,
			StreakBonus67: settings.StreakBonus67,
		}, streakBefore, now)
		loserMember.RegisterLoss(now)

		winnerLP := shared.LP(settings.LPWin) + lpBonus
		loserLP := shared.LP(settings.LPLoss)
		winnerMember.AddLP(winnerLP, now)
		loserMember.AddLP(loserLP, now)

		if err := uc.members.Update(ctx, winnerMember); err != nil {
			return fmt.Errorf("confirm result: update winner: %w", err)
		}
		if err := uc.members.Update(ctx, loserMember); err != nil {
			return fmt.Errorf("confirm result: update loser: %w", err)
		}

		// Бонус «каждая партия сверх t». Считаем по уже подтверждённым партиям.
		winnerTotal, _ := uc.matches.CountConfirmedByPlayerInSeason(ctx, m.SeasonID, winnerID)
		loserTotal, _ := uc.matches.CountConfirmedByPlayerInSeason(ctx, m.SeasonID, loserID)
		winnerStones := shared.Stones(settings.StonesWin) + shared.Stones(stoneBonus)
		loserStones := shared.Stones(settings.StonesLoss)
		if winnerTotal+1 > settings.T {
			winnerStones += shared.Stones(settings.StonesPerExtraMatch)
		}
		if loserTotal+1 > settings.T {
			loserStones += shared.Stones(settings.StonesPerExtraMatch)
		}

		// Обновляем игроков.
		winner, err := uc.players.GetByID(ctx, winnerID)
		if err != nil {
			return fmt.Errorf("confirm result: get winner: %w", err)
		}
		loser, err := uc.players.GetByID(ctx, loserID)
		if err != nil {
			return fmt.Errorf("confirm result: get loser: %w", err)
		}
		if err := winner.AddStones(winnerStones, now); err != nil {
			return fmt.Errorf("confirm result: winner stones: %w", err)
		}
		if err := loser.AddStones(loserStones, now); err != nil {
			return fmt.Errorf("confirm result: loser stones: %w", err)
		}
		winner.AddLP(winnerLP, now)
		loser.AddLP(loserLP, now)

		if err := uc.players.Update(ctx, winner); err != nil {
			return fmt.Errorf("confirm result: save winner: %w", err)
		}
		if err := uc.players.Update(ctx, loser); err != nil {
			return fmt.Errorf("confirm result: save loser: %w", err)
		}

		// Сохраняем партию.
		if err := uc.matches.Update(ctx, m); err != nil {
			return fmt.Errorf("confirm result: save match: %w", err)
		}

		// Пишем историю камней.
		reason := fmt.Sprintf("match:%s", m.ID)
		if err := uc.stones.Create(ctx, economy.NewTransaction(winnerID, winnerStones, reason, nil, now)); err != nil {
			return fmt.Errorf("confirm result: stone tx winner: %w", err)
		}
		if loserStones > 0 {
			if err := uc.stones.Create(ctx, economy.NewTransaction(loserID, loserStones, reason, nil, now)); err != nil {
				return fmt.Errorf("confirm result: stone tx loser: %w", err)
			}
		}

		// Событие и уведомление.
		if err := uc.events.Create(ctx, dommatch.NewEvent(m.ID, dommatch.EventConfirmed, in.ActorID, nil, now)); err != nil {
			return fmt.Errorf("confirm result: save event: %w", err)
		}
		if m.ResultSubmittedBy != nil {
			n := notification.New(*m.ResultSubmittedBy, notification.TypeResultConfirmed, map[string]any{
				"match_id": int64(m.ID),
			}, notification.ChannelEmail, now)
			if err := uc.notifs.Create(ctx, n); err != nil {
				return fmt.Errorf("confirm result: notify: %w", err)
			}
		}

		out = &dto.ConfirmResultOutput{
			Match:        m,
			WinnerLP:     winnerLP,
			LoserLP:      loserLP,
			WinnerStones: winnerStones,
			LoserStones:  loserStones,
			StreakBroken: lpBonus > 0,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
