package shared

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ---- Typed IDs ---------------------------------------------------------
//
// ID назначает репозиторий при Create (через RETURNING id в Postgres).
// Нулевое значение означает «ещё не сохранено».

type PlayerID int64
type AdminID int64
type ClubID int64
type SeasonID int64
type DivisionID int64
type DivisionMemberID int64
type MatchID int64
type MatchEventID int64
type BadgeID int64
type TitleID int64
type PlayerBadgeID int64
type ClubBadgeID int64
type SuperTitleID int64
type StoneTransactionID int64
type PurchaseID int64
type NotificationID int64

func (id PlayerID) String() string           { return strconv.FormatInt(int64(id), 10) }
func (id AdminID) String() string            { return strconv.FormatInt(int64(id), 10) }
func (id ClubID) String() string             { return strconv.FormatInt(int64(id), 10) }
func (id SeasonID) String() string           { return strconv.FormatInt(int64(id), 10) }
func (id DivisionID) String() string         { return strconv.FormatInt(int64(id), 10) }
func (id DivisionMemberID) String() string   { return strconv.FormatInt(int64(id), 10) }
func (id MatchID) String() string            { return strconv.FormatInt(int64(id), 10) }
func (id MatchEventID) String() string       { return strconv.FormatInt(int64(id), 10) }
func (id BadgeID) String() string            { return strconv.FormatInt(int64(id), 10) }
func (id TitleID) String() string            { return strconv.FormatInt(int64(id), 10) }
func (id PlayerBadgeID) String() string      { return strconv.FormatInt(int64(id), 10) }
func (id ClubBadgeID) String() string        { return strconv.FormatInt(int64(id), 10) }
func (id SuperTitleID) String() string       { return strconv.FormatInt(int64(id), 10) }
func (id StoneTransactionID) String() string { return strconv.FormatInt(int64(id), 10) }
func (id PurchaseID) String() string         { return strconv.FormatInt(int64(id), 10) }
func (id NotificationID) String() string     { return strconv.FormatInt(int64(id), 10) }

// IsZero — ID ещё не назначен (запись не сохранена).
func (id PlayerID) IsZero() bool { return id == 0 }

// ... аналогично при необходимости.

// ---- Общие ошибки (без изменений) -------------------------------------

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

// ---- Value objects (без изменений) ------------------------------------

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

type Nickname string

func NewNickname(raw string) (Nickname, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 2 || len(raw) > 32 {
		return "", ErrValidation
	}
	return Nickname(raw), nil
}

func (n Nickname) String() string { return string(n) }

type Rating int

func NewRating(v int) (Rating, error) {
	if v < 0 {
		return 0, ErrValidation
	}
	return Rating(v), nil
}
func (r Rating) IsZero() bool { return r == 0 }

type LP float64
type Stones float64

type Timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}
