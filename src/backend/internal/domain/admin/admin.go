package admin

import (
	"time"

	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// Admin — отдельная сущность. Админ не играет партии, у него отдельный вход.
type Admin struct {
	ID           shared.AdminID
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

func New(login, passwordHash string, now time.Time) (*Admin, error) {
	if login == "" || passwordHash == "" {
		return nil, shared.ErrValidation
	}
	return &Admin{
		ID:           shared.NewAdminID(),
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}, nil
}
