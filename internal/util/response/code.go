package response

type (
	ErrorCode   string
	SuccessCode string
)

// ============================
// Response Code Mapping
// ============================

// ───────────────
// General (Prefix 0): 00001++
// - General success/errors
// - Middleware errors
// - Database issues
// ───────────────

const (
	SuccessOK              SuccessCode = "00001"
	ErrInternalServerError ErrorCode   = "00002"
	ErrEmptyRequestBody    ErrorCode   = "00003"
	ErrDatabaseError       ErrorCode   = "00004"
	ErrUnprocessableEntity ErrorCode   = "00005"
)

// ───────────────
// Auth (Prefix 1): 10001++
// - Authentication & Authorization
// ───────────────

const (
	SuccessLogin          SuccessCode = "10001"
	ErrEmptyEmail         ErrorCode   = "10002"
	ErrInvalidEmailFormat ErrorCode   = "10003"
	ErrEmptyPassword      ErrorCode   = "10004"
	ErrInvalidCredentials ErrorCode   = "10005"

	ErrUnauthorized  ErrorCode = "10006"
	ErrTokenNotFound ErrorCode = "10007"
	ErrInvalidToken  ErrorCode = "10008"
)

// ───────────────
// User (Prefix 2): 20001++
// - User management
// ───────────────

const (
	SuccessUserUpdated SuccessCode = "20001"
	ErrUserNotFound    ErrorCode   = "20002"
)

// ───────────────
// Customer (Prefix 3): 30001++
// - Customer management
// ───────────────

const (
	SuccessCustomerCreated SuccessCode = "30001"
	ErrCustomerNotFound    ErrorCode   = "30002"
)
