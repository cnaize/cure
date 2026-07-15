package adapter

import "io"

type Adapter interface {
	Check(name string) bool
	Adapt(in io.Reader) io.Reader
}
