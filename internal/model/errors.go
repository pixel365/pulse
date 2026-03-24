package model

type ErrorKind string

var (
	ErrNone         ErrorKind = ""
	ErrTimeout      ErrorKind = "timeout"
	ErrNetwork      ErrorKind = "network"
	ErrStatusCode   ErrorKind = "status_code"
	ErrResponseBody ErrorKind = "response_body"
	ErrUnknown      ErrorKind = "unknown"
)
