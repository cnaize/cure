package adapter

import (
	"io"
	"path/filepath"
	"slices"
	"strings"
)

var yaraFileExtensions = []string{".yar", ".yara"}

var _ Adapter = (*Yara)(nil)

type Yara struct {
	include []string
	exclude []string
}

func NewYara() *Yara {
	return &Yara{}
}

func (a *Yara) WithInclude(names ...string) *Yara {
	a.include = names

	return a
}

func (a *Yara) WithExclude(names ...string) *Yara {
	a.exclude = names

	return a
}

func (a *Yara) Check(name string) bool {
	name = strings.ToLower(name)
	if !slices.Contains(yaraFileExtensions, filepath.Ext(name)) {
		return false
	}

	if len(a.include) > 0 {
		return slices.ContainsFunc(a.include, func(file string) bool {
			return strings.HasSuffix(name, file)
		})
	}

	if len(a.exclude) > 0 {
		return !slices.ContainsFunc(a.exclude, func(file string) bool {
			return strings.HasSuffix(name, file)
		})
	}

	return true
}

func (a *Yara) Adapt(name string, in io.Reader) io.Reader {
	return in
}
