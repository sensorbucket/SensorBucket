package cleanupper

import (
	"context"
	"errors"
	"time"
)

type Shutdown func(context.Context) error

type Cleanupper []Shutdown

func Create() Cleanupper {
	return make(Cleanupper, 0)
}

func (c *Cleanupper) Add(fn Shutdown) {
	*c = append(*c, fn)
}

func (c *Cleanupper) Execute(timeout time.Duration) error {
	ctxTO, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cleanupErrors error
	for _, fn := range *c {
		cleanupErrors = errors.Join(cleanupErrors, fn(ctxTO))
	}

	return cleanupErrors
}
