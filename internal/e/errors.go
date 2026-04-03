package e

import (
	"context"
	"errors"
	"net"
)

type ErrorKind string

func (e ErrorKind) String() string {
	return string(e)
}

const (
	// ErrNone means the execution completed without an error.
	ErrNone ErrorKind = ""

	// ErrTimeout covers context deadlines and operation timeouts.
	ErrTimeout ErrorKind = "timeout"

	// ErrNetwork covers connection establishment and transport-level failures.
	ErrNetwork ErrorKind = "network"

	// ErrProtocol means the endpoint responded, but the protocol-level response was invalid.
	ErrProtocol ErrorKind = "protocol"

	// ErrConstraint means the endpoint responded, but the observed result did not satisfy check expectations.
	ErrConstraint ErrorKind = "constraint"

	// ErrInternal is used for local checker/runtime errors unrelated to the target system.
	ErrInternal ErrorKind = "internal"

	// ErrUnknown is a fallback for errors that were not classified more precisely.
	ErrUnknown ErrorKind = "unknown"
)

var (
	ErrNotFound = errors.New("not found")
)

type KindError struct {
	Err  error
	Kind ErrorKind
}

func (e *KindError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}

	return e.Err.Error()
}

func (e *KindError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func NewError(kind ErrorKind, msg string) error {
	return &KindError{
		Kind: kind,
		Err:  errors.New(msg),
	}
}

func ResolveError(err error) (ErrorKind, string) {
	if err == nil {
		return ErrNone, ""
	}

	if kindErr, ok := errors.AsType[*KindError](err); ok && kindErr != nil {
		return kindErr.Kind, err.Error()
	}

	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return ErrTimeout, err.Error()
	}

	if netErr, ok := errors.AsType[net.Error](err); ok {
		if netErr.Timeout() {
			return ErrTimeout, err.Error()
		}

		return ErrNetwork, err.Error()
	}

	if _, ok := errors.AsType[*net.OpError](err); ok {
		return ErrNetwork, err.Error()
	}

	return ErrUnknown, err.Error()
}
