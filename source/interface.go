package source

import (
	"context"
	"io"
	"iter"
)

type Source interface {
	Files(ctx context.Context) (iter.Seq[io.Reader], error)
}
