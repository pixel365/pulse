package http

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/pixel365/pulse/internal/checker"
	c "github.com/pixel365/pulse/internal/config"
)

type Alias = c.TypedCheck[c.HttpSpec]

var _ checker.Checker = (*Checker)(nil)

type Checker struct {
	config Alias
}

func NewChecker(cfg Alias) *Checker {
	return &Checker{cfg}
}

func (c *Checker) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	sleep(ctx, c.config.Jitter)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			sleep(ctx, c.config.Jitter)

			err := c.execute(ctx)
			//TODO: handle errors
			switch {
			case errors.Is(err, context.Canceled):
			case errors.Is(err, context.DeadlineExceeded):
			case errors.Is(err, ErrCode):
			case errors.Is(err, ErrResponseBody):
			case errors.Is(err, ErrCtxCancelled):
			case errors.Is(err, ErrTimeout):
			default:
			}
		}
	}
}

func sleep(ctx context.Context, t time.Duration) {
	if t <= 0 {
		return
	}

	val, err := jitter(t)
	if err != nil {
		val = 0
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(val):
	}
}

func jitter(max time.Duration) (time.Duration, error) {
	if max <= 0 {
		return 0, nil
	}

	n, err := rand.Int(rand.Reader, big.NewInt(max.Nanoseconds()+1))
	if err != nil {
		return 0, err
	}

	return time.Duration(n.Int64()), nil
}
