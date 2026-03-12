package dns

import (
	"context"

	"github.com/pixel365/pulse/internal/checker"
	c "github.com/pixel365/pulse/internal/config"
)

type Alias = c.TypedCheck[c.DNSSpec]

var _ checker.Checker = (*Checker)(nil)

type Checker struct {
	config Alias
}

func NewChecker(cfg Alias) *Checker {
	return &Checker{cfg}
}

func (c *Checker) Run(ctx context.Context) error {
	return nil
}
