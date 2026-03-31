package internal

import (
	"context"
)

type Checker interface {
	Check(context.Context) error
}
