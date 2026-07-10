package middleware

import "time"

type ScanNeeded int

const (
	ScanNeededBody ScanNeeded = 1 << iota
	ScanNeededQuery
	ScanNeededHeaders
	ScanNeededCookies

	ScanNeededFull = ScanNeededBody | ScanNeededQuery | ScanNeededHeaders | ScanNeededCookies
)

type Options struct {
	ScanNeeded  ScanNeeded
	ScanTimeout time.Duration
	MaxBuffSize int
}

func (o *Options) SetDefaults() {
	if o.ScanNeeded < 1 {
		o.ScanNeeded = ScanNeededBody | ScanNeededQuery
	}

	if o.ScanTimeout < 1 {
		o.ScanTimeout = 100 * time.Millisecond
	}

	if o.MaxBuffSize < 1 {
		o.MaxBuffSize = 10 * 1024 // 10KB
	}
}
