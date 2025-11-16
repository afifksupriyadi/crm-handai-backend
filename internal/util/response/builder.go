package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/danielgtaylor/huma/v2"
)

// BuildSuccess constructs a standardized success response object.
func BuildSuccess(data interface{}, code SuccessCode) *Response {
	if detail, found := getSuccessDetail(code); found {
		return &Response{
			Status: detail.HTTPStatus,
			Body:   buildResponseBody(string(code), detail.Message, data),
		}
	}

	return &Response{
		Status: http.StatusOK,
		Body:   buildResponseBody(string(code), "Unknown success code", data),
	}
}

// BuildError is used in handlers that return (*Response, error).
// It converts an application-level error into a standardized Response object.
func BuildError(ctx context.Context, err error) *Response {
	parsed := ParseErrorWithHTTP(err)
	return &Response{
		Status: parsed.HTTPCode,
		Body:   buildResponseBody(string(parsed.ErrCode), parsed.ErrMessage, struct{}{}),
	}
}

// WriteError is intended for use in middlewares or handlers that work with huma.Context
// and do not return values explicitly.
func WriteError(ctx huma.Context, err error) {
	parsed := ParseErrorWithHTTP(err)
	ctx.SetStatus(parsed.HTTPCode)
	ctx.SetHeader("Content-Type", "application/json")

	resp := buildResponseBody(string(parsed.ErrCode), parsed.ErrMessage, struct{}{})

	if err := json.NewEncoder(ctx.BodyWriter()).Encode(resp); err != nil {
		logger.FromContext(ctx.Context(), 1).Error().
			Err(err).
			Str("original_error", parsed.ErrMessage).
			Int("status_code", parsed.HTTPCode).
			Msg("Failed to encode error response")

		apiErr, _ := getErrorDetail(ErrInternalServerError)
		ctx.SetStatus(apiErr.HTTPStatus)
		_ = json.NewEncoder(ctx.BodyWriter()).Encode(map[string]interface{}{
			"data":    struct{}{},
			"code":    apiErr.Code,
			"message": apiErr.Message,
		})
	}
}

func buildResponseBody(code string, message string, data interface{}) Body {
	return Body{
		Data:    data,
		Code:    code,
		Message: message,
	}
}
