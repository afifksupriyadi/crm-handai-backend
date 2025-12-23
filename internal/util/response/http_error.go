package response

import "net/http"

// ParsedError represents an error that has been parsed for HTTP response
type ParsedError struct {
	HTTPCode   int
	ErrMessage string
	ErrCode    ErrorCode
}

// ParseErrorWithHTTP converts an error to a ParsedError for HTTP response
func ParseErrorWithHTTP(err error) *ParsedError {
	if appErr, ok := err.(*AppError); ok {
		if detail, found := getErrorDetail(appErr.ErrCode, appErr.Args...); found {
			return &ParsedError{detail.HTTPStatus, detail.Message, ErrorCode(detail.Code)}
		}
		return &ParsedError{http.StatusInternalServerError, "Unknown error", appErr.ErrCode}
	}

	return &ParsedError{http.StatusInternalServerError, "Internal Server Error", ErrInternalServerError}
}
