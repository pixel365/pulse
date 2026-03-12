package http

import "errors"

var (
	ErrTimeout      = errors.New("timeout")
	ErrCtxCancelled = errors.New("context cancelled")
	ErrCode         = errors.New("unsuccess code")
	ErrResponseBody = errors.New("invalid response body")
)
