package logger

import (
	"context"
	"log/slog"
	"os"
	"strconv"
)

var _ Logger = (*Slog)(nil)

type Slog struct {
	logger *slog.Logger
	debug  bool
}

func (l *Slog) Debug(ctx context.Context, msg string, args ...any) {
	if l == nil || !l.debug {
		return
	}

	l.logger.DebugContext(ctx, msg, args...)
}

func (l *Slog) Info(ctx context.Context, msg string, args ...any) {
	if l == nil {
		return
	}

	l.logger.InfoContext(ctx, msg, args...)
}

func (l *Slog) Warn(ctx context.Context, msg string, args ...any) {
	if l == nil {
		return
	}

	l.logger.WarnContext(ctx, msg, args...)
}

func (l *Slog) Error(ctx context.Context, msg string, args ...any) {
	if l == nil {
		return
	}

	l.logger.ErrorContext(ctx, msg, args...)
}

func NewSlog() *Slog {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	return &Slog{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		debug:  debug,
	}
}
