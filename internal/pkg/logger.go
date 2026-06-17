package pkg

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
)

// Logger wraps sentry.Logger to provide structured logging linked to traces.
// It must be created per-request using NewLogger(ctx) so logs are correlated
// with the current trace context.
type Logger struct {
	logger sentry.Logger
}

// NewLogger creates a Logger bound to the given context.
// Use the context from gin.Context.Request.Context() in handlers.
func NewLogger(ctx context.Context) *Logger {
	return &Logger{logger: sentry.NewLogger(ctx)}
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info().Emit(fmt.Sprintf(msg, args...))
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn().Emit(fmt.Sprintf(msg, args...))
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error().Emit(fmt.Sprintf(msg, args...))
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug().Emit(fmt.Sprintf(msg, args...))
}
