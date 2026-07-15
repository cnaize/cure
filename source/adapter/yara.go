package adapter

import (
	"io"
	"path/filepath"
	"slices"
	"strings"
)

var YaraFileExtensions = []string{".yar", ".yara"}

var _ Adapter = Yara{}

type Yara struct {
}

func NewYara() Yara {
	return Yara{}
}

func (a Yara) Check(name string) bool {
	return slices.Contains(YaraFileExtensions, strings.ToLower(filepath.Ext(name)))
}

func (a Yara) Adapt(in io.Reader) io.Reader {
	return in
}
