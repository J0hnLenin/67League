package ports

import (
	"context"

	"github.com/J0hnLenin/67League/internal/domain/notification"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// NotificationRepository — хранение уведомлений.
type NotificationRepository interface {
	Create(ctx context.Context, n *notification.Notification) error
	GetByID(ctx context.Context, id shared.NotificationID) (*notification.Notification, error)
	Update(ctx context.Context, n *notification.Notification) error

	ListByPlayer(ctx context.Context, playerID shared.PlayerID, p Pagination) ([]*notification.Notification, error)

	// ListUnsent — уведомления, которые ещё не отправлены (для воркера отправки).
	ListUnsent(ctx context.Context, limit int) ([]*notification.Notification, error)
}
