// Package migrations встраивает SQL-миграции в бинарь,
// чтобы их можно было применять без файловой системы.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
