package source

import (
	"context"
	"io"
	"iter"
)

var YaraFileExtensions = []string{".yar", ".yara"}

type Source interface {
	Files(ctx context.Context) (iter.Seq[io.Reader], error)
}
