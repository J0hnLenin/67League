package settings

import (
	"time"
)

// SiteSettings — ненормализованное key-value хранилище всех глобальных констант.
type SiteSettings map[string]any

type Entry struct {
	Key       string
	Value     any
	UpdatedAt time.Time
}
