package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// RegisterHumaErrorHandler overrides huma.NewError to provide custom error responses
// for validation errors (HTTP 422). Should be called once during application startup.
// Note: Overriding huma.NewError is not thread-safe if called multiple times.
func RegisterHumaErrorHandler() {
	defaultHumaError := huma.NewError
	huma.NewError = func(status int, message string, errs ...error) huma.StatusError {
		if status != http.StatusUnprocessableEntity {
			return defaultHumaError(status, message, errs...)
		}

		msg := buildUnprocessableEntityMessage(errs)
		if msg == "" {
			msg = "Terjadi kesalahan pada data yang dikirim"
		}

		return &HumaError{
			Body: Body{
				Code:    string(ErrUnprocessableEntity),
				Message: msg,
				Data:    struct{}{},
			},
			Status: http.StatusBadRequest,
		}
	}
}

// buildUnprocessableEntityMessage extracts human-readable messages from Huma error details.
func buildUnprocessableEntityMessage(errs []error) string {
	var messages []string
	for _, err := range errs {
		if em, ok := err.(*huma.ErrorDetail); ok && em.Message != "" && em.Location != "" {
			messages = append(messages, fmt.Sprintf("%s (%s)", em.Message, em.Location))
		}
	}
	return strings.Join(messages, "; ")
}
