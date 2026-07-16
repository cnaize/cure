package adapter

import "io"

type Adapter interface {
	Check(name string) bool
	Adapt(name string, in io.Reader) io.Reader
}
