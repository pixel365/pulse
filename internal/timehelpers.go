package internal

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"
)

func Sleep(ctx context.Context, t time.Duration) {
	if t <= 0 {
		return
	}

	val, err := Jitter(t)
	if err != nil {
		val = 0
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(val):
	}
}

func Jitter(max time.Duration) (time.Duration, error) {
	if max <= 0 {
		return 0, nil
	}

	n, err := rand.Int(rand.Reader, big.NewInt(max.Nanoseconds()+1))
	if err != nil {
		return 0, err
	}

	return time.Duration(n.Int64()), nil
}
