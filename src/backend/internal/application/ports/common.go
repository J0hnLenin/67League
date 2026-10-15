// Package ports содержит интерфейсы, которые бизнес-логика требует
// от внешнего мира (БД, почта, S3, внешние API, время и т.д.).
//
// Реализации живут в internal/infrastructure.
// Домен и application не знают об их конкретных реализациях.
package ports

// Pagination — общая пагинация для списочных запросов.
type Pagination struct {
	Limit  int
	Offset int
}

// DefaultPagination возвращает разумные значения по умолчанию.
func DefaultPagination() Pagination {
	return Pagination{Limit: 50, Offset: 0}
}

// Normalize приводит значения к допустимым.
func (p Pagination) Normalize() Pagination {
	if p.Limit <= 0 || p.Limit > 500 {
		p.Limit = 50
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}

// SortDirection — направление сортировки.
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)
