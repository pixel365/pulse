package checker

import (
	"context"
)

type Checker interface {
	Run(context.Context) error
}
