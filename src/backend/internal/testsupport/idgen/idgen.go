// internal/testsupport/idgen/idgen.go
package idgen

import "sync/atomic"

type Gen struct{ n atomic.Int64 }

func New(start int64) *Gen { g := &Gen{}; g.n.Store(start); return g }

func (g *Gen) Next() int64 { return g.n.Add(1) }
