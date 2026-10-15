package ports

import "context"

// TransactionManager — управление транзакциями БД на уровне use case.
//
// Реализация (например, для pgx) извлекает активное соединение из ctx и
// передаёт его во все репозитории. Репозитории сами решают, использовать
// ли tx из ctx или открыть новое соединение.
//
// Пример использования в use case:
//
//	err := tx.WithinTransaction(ctx, func(ctx context.Context) error {
//	    if err := players.Create(ctx, p); err != nil { return err }
//	    return matches.Create(ctx, m)
//	})
type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
