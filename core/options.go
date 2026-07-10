package core

import "time"

type Options struct {
	UpdateInterval time.Duration
}

func (o *Options) SetDefaults() {
	if o.UpdateInterval < 1 {
		o.UpdateInterval = 24 * time.Hour
	}
}
