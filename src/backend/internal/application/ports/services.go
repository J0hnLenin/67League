package ports

import (
	"context"
	"io"
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// Clock — абстракция времени. Нужна, чтобы use cases были детерминированы в тестах.
type Clock interface {
	Now() time.Time
}

// PasswordHasher — хеширование и проверка паролей.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(plain, hash string) (bool, error)
}

// TokenSubject — кто предъявитель токена.
type TokenSubject string

const (
	SubjectPlayer TokenSubject = "player"
	SubjectAdmin  TokenSubject = "admin"
)

// AuthClaims — полезная нагрузка токена.
type AuthClaims struct {
	Subject   TokenSubject
	UserID    string // string(UUID), чтобы не тащить доменные типы в инфраструктуру
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// TokenIssuer — выпуск и разбор токенов (JWT или серверные сессии — не важно).
type TokenIssuer interface {
	IssuePlayer(playerID shared.PlayerID) (string, error)
	IssueAdmin(adminID shared.AdminID) (string, error)
	Parse(token string) (*AuthClaims, error)
}

// EmailSender — отправка писем.
type EmailSender interface {
	// Send — простое письмо.
	Send(ctx context.Context, to shared.Email, subject, body string) error

	// SendTemplate — письмо по шаблону.
	// template — код шаблона (например, "new_challenge"),
	// data — параметры шаблона.
	SendTemplate(ctx context.Context, to shared.Email, template string, data map[string]any) error
}

// Storage — S3-совместимое хранилище для SGF и изображений.
type Storage interface {
	// Put сохраняет данные и возвращает ключ.
	Put(ctx context.Context, key string, data io.Reader, contentType string) error

	// Delete удаляет объект по ключу.
	Delete(ctx context.Context, key string) error

	// URL возвращает публичный URL объекта.
	URL(ctx context.Context, key string) (string, error)
}

// ExternalRating — рейтинг, полученный из внешнего источника.
type ExternalRating struct {
	RFGID  string
	Nick   string
	Rating shared.Rating
	Source string // "rfg" | "sputnik"
}

// RatingProvider — клиент внешних источников рейтинга (РФГ, Спутник).
// Используется при регистрации игрока для автоподтягивания рейтинга.
type RatingProvider interface {
	FetchByRFGID(ctx context.Context, rfgID string) (*ExternalRating, error)
	FetchByRFGNick(ctx context.Context, nick string) (*ExternalRating, error)
	FetchBySputnikID(ctx context.Context, sputnikID string) (*ExternalRating, error)
}

// OGSClient — клиент OGS API (для автоимпорта партий, фаза 3).
// Пока заглушка — реализация появится позже.
type OGSClient interface {
	FetchGame(ctx context.Context, gameID string) (*OGSGame, error)
	ListPlayerGames(ctx context.Context, ogsNick string, since time.Time) ([]*OGSGame, error)
}

type OGSGame struct {
	ID         string
	BlackNick  string
	WhiteNick  string
	WinnerNick string
	ScoreDiff  float64
	SGFURL     string
	PlayedAt   time.Time
}
