package response

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/caller"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
)

// AppError represents an application-specific error with additional context
type AppError struct {
	OriginalError error
	ErrCode       ErrorCode
	Location      string
	Args          []interface{}
}

// Error implements the error interface
func (ae *AppError) Error() string {
	if ae.OriginalError != nil {
		return ae.OriginalError.Error()
	}
	return "unknown error"
}

// WrapAppError creates a new AppError with the given parameters and logs it
func WrapAppError(ctx context.Context, originalError error, code ErrorCode, msgCustom string, args ...interface{}) *AppError {
	appErr := &AppError{
		OriginalError: originalError,
		ErrCode:       code,
		Location:      caller.GetWithDepth(2),
		Args:          args,
	}

	LogAppError(ctx, appErr, msgCustom)

	return appErr
}

// LogAppError logs the application error with context
func LogAppError(ctx context.Context, appErr *AppError, msgCustom string) {
	msg := msgCustom
	if appErr.OriginalError != nil {
		msg = msgCustom + ": " + appErr.OriginalError.Error()
	}

	logger.FromContext(ctx, 1).Error().
		Interface("errorDetail", map[string]interface{}{
			"location": appErr.Location,
			"message":  msg,
		}).
		Msg("Error occurred")
}
