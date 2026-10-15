package ports

import (
	"context"

	"github.com/J0hnLenin/67League/internal/domain/settings"
)

// SettingsRepository — глобальные настройки сайта (key-value, JSONB).
type SettingsRepository interface {
	Get(ctx context.Context, key string) (any, error)
	GetAll(ctx context.Context) (settings.SiteSettings, error)
	Set(ctx context.Context, key string, value any) error
	SetMany(ctx context.Context, values settings.SiteSettings) error
}
