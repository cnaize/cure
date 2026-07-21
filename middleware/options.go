package middleware

import "time"

type ScanNeeded int

const (
	ScanNeededBody ScanNeeded = 1 << iota
	ScanNeededPath
	ScanNeededQueryKey
	ScanNeededQueryVal
	ScanNeededHeadersKey
	ScanNeededHeadersVal
	ScanNeededCookiesKey
	ScanNeededCookiesVal

	ScanNeededFull = (1 << iota) - 1

	ScanNeededQuery   = ScanNeededQueryKey | ScanNeededQueryVal
	ScanNeededHeaders = ScanNeededHeadersKey | ScanNeededHeadersVal
	ScanNeededCookies = ScanNeededCookiesKey | ScanNeededCookiesVal
)

type Options struct {
	ScanNeeded  ScanNeeded
	ScanTimeout time.Duration
	MaxBuffSize int
}

func (o *Options) SetDefaults() {
	if o.ScanNeeded < 1 {
		o.ScanNeeded = ScanNeededFull
	}

	if o.ScanTimeout < 1 {
		o.ScanTimeout = 10 * time.Millisecond
	}

	if o.MaxBuffSize < 1 {
		o.MaxBuffSize = 10 * 1024 // 10KB
	}
}
