package player

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// RegisterPlayer регистрирует нового игрока.
//
// Побочные эффекты (все внутри одной транзакции):
//   - создаётся Player;
//   - если задан ClubID — игрок вступает в клуб;
//   - если есть активный сезон — игрок попадает в дивизион по рейтингу.
type RegisterPlayer struct {
	players   ports.PlayerRepository
	clubs     ports.ClubRepository
	seasons   ports.SeasonRepository
	divisions ports.DivisionRepository
	members   ports.DivisionMemberRepository
	rating    ports.RatingProvider
	hasher    ports.PasswordHasher
	clock     ports.Clock
	tx        ports.TransactionManager
}

type RegisterPlayerDeps struct {
	Players   ports.PlayerRepository
	Clubs     ports.ClubRepository
	Seasons   ports.SeasonRepository
	Divisions ports.DivisionRepository
	Members   ports.DivisionMemberRepository
	Rating    ports.RatingProvider
	Hasher    ports.PasswordHasher
	Clock     ports.Clock
	Tx        ports.TransactionManager
}

func NewRegisterPlayer(d RegisterPlayerDeps) (*RegisterPlayer, error) {
	if d.Players == nil || d.Clubs == nil || d.Seasons == nil ||
		d.Divisions == nil || d.Members == nil || d.Hasher == nil ||
		d.Clock == nil || d.Tx == nil {
		return nil, errors.New("register player: missing dependencies")
	}
	return &RegisterPlayer{
		players:   d.Players,
		clubs:     d.Clubs,
		seasons:   d.Seasons,
		divisions: d.Divisions,
		members:   d.Members,
		rating:    d.Rating,
		hasher:    d.Hasher,
		clock:     d.Clock,
		tx:        d.Tx,
	}, nil
}

func (uc *RegisterPlayer) Execute(ctx context.Context, in dto.RegisterPlayerInput) (*dto.RegisterPlayerOutput, error) {
	if err := uc.validate(ctx, in); err != nil {
		return nil, err
	}

	rating := uc.resolveRating(ctx, in)

	hash, err := uc.hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("register player: hash password: %w", err)
	}

	now := uc.clock.Now()
	p, err := player.New(player.NewPlayerParams{
		FullName:     in.FullName,
		Email:        in.Email,
		PasswordHash: hash,
		Nickname:     in.Nickname,
		City:         in.City,
		Rating:       rating,
		KGSNick:      in.KGSNick,
		OGSNick:      in.OGSNick,
		FoxNick:      in.FoxNick,
		RFGID:        in.RFGID,
		SputnikID:    in.SputnikID,
	}, now)
	if err != nil {
		return nil, fmt.Errorf("register player: build entity: %w", err)
	}

	var assignedDivision *shared.DivisionID

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := uc.players.Create(ctx, p); err != nil {
			return fmt.Errorf("create player: %w", err)
		}

		if in.ClubID != nil {
			if err := uc.attachToClub(ctx, p, *in.ClubID); err != nil {
				return err
			}
		}

		div, err := uc.assignDivision(ctx, p, now)
		if err != nil {
			return err
		}
		assignedDivision = div
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.RegisterPlayerOutput{Player: p, Division: assignedDivision}, nil
}

func (uc *RegisterPlayer) validate(ctx context.Context, in dto.RegisterPlayerInput) error {
	if in.Email == "" {
		return fmt.Errorf("register player: email: %w", shared.ErrValidation)
	}
	if in.Nickname == "" {
		return fmt.Errorf("register player: nickname: %w", shared.ErrValidation)
	}
	if len(in.Password) < 6 {
		return fmt.Errorf("register player: password too short: %w", shared.ErrValidation)
	}

	exists, err := uc.players.ExistsByEmail(ctx, in.Email)
	if err != nil {
		return fmt.Errorf("register player: check email: %w", err)
	}
	if exists {
		return fmt.Errorf("register player: email %q: %w", in.Email, shared.ErrAlreadyExists)
	}

	exists, err = uc.players.ExistsByNickname(ctx, in.Nickname)
	if err != nil {
		return fmt.Errorf("register player: check nickname: %w", err)
	}
	if exists {
		return fmt.Errorf("register player: nickname %q: %w", in.Nickname, shared.ErrAlreadyExists)
	}
	return nil
}

// resolveRating пробует подтянуть рейтинг из внешнего источника.
// Если не удалось — рейтинг 0. Ошибки внешнего источника не критичны.
func (uc *RegisterPlayer) resolveRating(ctx context.Context, in dto.RegisterPlayerInput) shared.Rating {
	if uc.rating == nil {
		return 0
	}
	if in.RFGID != "" {
		if r, err := uc.rating.FetchByRFGID(ctx, in.RFGID); err == nil && r != nil {
			return r.Rating
		}
	}
	if in.SputnikID != "" {
		if r, err := uc.rating.FetchBySputnikID(ctx, in.SputnikID); err == nil && r != nil {
			return r.Rating
		}
	}
	if in.Nickname != "" {
		if r, err := uc.rating.FetchByRFGNick(ctx, string(in.Nickname)); err == nil && r != nil {
			return r.Rating
		}
	}
	return 0
}

func (uc *RegisterPlayer) attachToClub(ctx context.Context, p *player.Player, clubID shared.ClubID) error {
	if _, err := uc.clubs.GetByID(ctx, clubID); err != nil {
		return fmt.Errorf("attach to club: %w", err)
	}
	if err := p.JoinClub(clubID, uc.clock.Now()); err != nil {
		return fmt.Errorf("attach to club: %w", err)
	}
	return uc.players.Update(ctx, p)
}

// assignDivision помещает игрока в дивизион активного сезона.
// Возвращает nil, если сейчас нет сезона, доступного для вступления.
func (uc *RegisterPlayer) assignDivision(ctx context.Context, p *player.Player, now time.Time) (*shared.DivisionID, error) {
	s, err := uc.seasons.GetCurrent(ctx)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("assign division: get season: %w", err)
	}
	if !s.Status.CanJoin() {
		return nil, nil
	}

	div, err := uc.divisions.FindForRating(ctx, s.ID, p.Rating)
	if err != nil {
		if !errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("assign division: find for rating: %w", err)
		}
		div, err = uc.divisions.GetLowest(ctx, s.ID)
		if err != nil {
			// Сезон есть, но дивизионов нет — это ошибка конфигурации.
			return nil, fmt.Errorf("assign division: get lowest: %w", err)
		}
	}

	m := division.NewMember(div.ID, p.ID, now)
	if err := uc.members.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("assign division: create member: %w", err)
	}
	return &div.ID, nil
}
