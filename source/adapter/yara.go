package adapter

import (
	"io"
	"path/filepath"
	"slices"
	"strings"
)

var yaraFileExtensions = []string{".yar", ".yara"}

var _ Adapter = Yara{}

type Yara struct {
}

func NewYara() Yara {
	return Yara{}
}

func (a Yara) Check(name string) bool {
	return slices.Contains(yaraFileExtensions, strings.ToLower(filepath.Ext(name)))
}

func (a Yara) Adapt(name string, in io.Reader) io.Reader {
	return in
}
