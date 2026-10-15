// Package shared содержит базовые типы, ID и ошибки, общие для всего домена.
package shared

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ---- Typed IDs ---------------------------------------------------------

type PlayerID uuid.UUID
type AdminID uuid.UUID
type ClubID uuid.UUID
type SeasonID uuid.UUID
type DivisionID uuid.UUID
type DivisionMemberID uuid.UUID
type MatchID uuid.UUID
type MatchEventID uuid.UUID
type BadgeID uuid.UUID
type TitleID uuid.UUID
type PlayerBadgeID uuid.UUID
type ClubBadgeID uuid.UUID
type SuperTitleID uuid.UUID
type StoneTransactionID uuid.UUID
type PurchaseID uuid.UUID
type NotificationID uuid.UUID

func NewPlayerID() PlayerID                     { return PlayerID(uuid.New()) }
func NewAdminID() AdminID                       { return AdminID(uuid.New()) }
func NewClubID() ClubID                         { return ClubID(uuid.New()) }
func NewSeasonID() SeasonID                     { return SeasonID(uuid.New()) }
func NewDivisionID() DivisionID                 { return DivisionID(uuid.New()) }
func NewDivisionMemberID() DivisionMemberID     { return DivisionMemberID(uuid.New()) }
func NewMatchID() MatchID                       { return MatchID(uuid.New()) }
func NewMatchEventID() MatchEventID             { return MatchEventID(uuid.New()) }
func NewBadgeID() BadgeID                       { return BadgeID(uuid.New()) }
func NewTitleID() TitleID                       { return TitleID(uuid.New()) }
func NewPlayerBadgeID() PlayerBadgeID           { return PlayerBadgeID(uuid.New()) }
func NewClubBadgeID() ClubBadgeID               { return ClubBadgeID(uuid.New()) }
func NewSuperTitleID() SuperTitleID             { return SuperTitleID(uuid.New()) }
func NewStoneTransactionID() StoneTransactionID { return StoneTransactionID(uuid.New()) }
func NewPurchaseID() PurchaseID                 { return PurchaseID(uuid.New()) }
func NewNotificationID() NotificationID         { return NotificationID(uuid.New()) }

// ---- Общие ошибки ------------------------------------------------------

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidState  = errors.New("invalid state")
	ErrForbidden     = errors.New("forbidden")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrValidation    = errors.New("validation error")
	ErrConflict      = errors.New("conflict")
	ErrLimitExceeded = errors.New("limit exceeded")
)

// ---- Value objects -----------------------------------------------------

// Email — валидированный email.
type Email string

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func NewEmail(raw string) (Email, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if !emailRe.MatchString(raw) {
		return "", ErrValidation
	}
	return Email(raw), nil
}

func (e Email) String() string { return string(e) }

// Nickname — игровой ник в лиге.
type Nickname string

func NewNickname(raw string) (Nickname, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 2 || len(raw) > 32 {
		return "", ErrValidation
	}
	return Nickname(raw), nil
}

func (n Nickname) String() string { return string(n) }

// Rating — рейтинг РФГ. 0 означает «нет рейтинга».
type Rating int

func NewRating(v int) (Rating, error) {
	if v < 0 {
		return 0, ErrValidation
	}
	return Rating(v), nil
}

func (r Rating) IsZero() bool { return r == 0 }

// LP — League Points, соревновательный рейтинг внутри сезона. Не тратится.
type LP float64

// Stones — игровая валюта. Может быть дробной из-за бонуса +0.5.
type Stones float64

// ---- Timestamps --------------------------------------------------------

type Timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}
