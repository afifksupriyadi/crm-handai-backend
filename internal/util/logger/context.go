package logger

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// FromContext returns a logger with values from context
func FromContext(ctx context.Context, skip int) *zerolog.Logger {
	if ctx == nil {
		ctx = context.Background()
	}

	logger := log.Ctx(ctx)

	// if no logger in context, use global logger
	if logger.GetLevel() == zerolog.Disabled {
		logger = &log.Logger
	}

	// apply caller skip if needed
	if skip > 0 {
		l := logger.With().CallerWithSkipFrameCount(skip + 2).Logger()
		return &l
	}

	return logger
}

// WithRequestID adds request_id to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	logger := log.Ctx(ctx).With().Str("request_id", requestID).Logger()
	return logger.WithContext(ctx)
}
