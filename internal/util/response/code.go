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
// - Health & Welcome
// ───────────────

const (
	SuccessOK              SuccessCode = "00001"
	SuccessWelcome         SuccessCode = "00002"
	SuccessHealthCheck     SuccessCode = "00003"
	ErrInternalServerError ErrorCode   = "00004"
	ErrEmptyRequestBody    ErrorCode   = "00005"
	ErrDatabaseError       ErrorCode   = "00006"
	ErrUnprocessableEntity ErrorCode   = "00007"
)

// ───────────────
// Auth (Prefix 1): 10001++
// - Authentication & Authorization
// ───────────────

const (
	SuccessLogin SuccessCode = "10001"

	ErrEmptyEmail         ErrorCode = "10002"
	ErrInvalidEmailFormat ErrorCode = "10003"
	ErrEmptyPassword      ErrorCode = "10004"
	ErrInvalidCredentials ErrorCode = "10005"
	ErrUnauthorized       ErrorCode = "10006"
	ErrTokenNotFound      ErrorCode = "10007"
	ErrInvalidToken       ErrorCode = "10008"
)

// ───────────────
// User (Prefix 2): 20001++
// - User management
// ───────────────

const (
	SuccessUserUpdated SuccessCode = "20001"

	ErrUserNotFound ErrorCode = "20002"
)

// ───────────────
// Customer (Prefix 3): 30001++
// - Customer management
// ───────────────

const (
	SuccessCustomerCreated   SuccessCode = "30001"
	SuccessCustomerRetrieved SuccessCode = "30002"
	SuccessCustomerUpdated   SuccessCode = "30003"
	SuccessCustomerDeleted   SuccessCode = "30004"

	ErrCustomerNotFound   ErrorCode = "30005"
	ErrEmptyName          ErrorCode = "30006"
	ErrNameTooLong        ErrorCode = "30007"
	ErrEmptyPhone         ErrorCode = "30008"
	ErrInvalidPhoneFormat ErrorCode = "30009"
	ErrPhoneTooLong       ErrorCode = "30010"
	ErrPhoneAlreadyExists ErrorCode = "30011"
)

// ───────────────
// Product (Prefix 4): 40001++
// - Product management
// ───────────────

const (
	SuccessProductCreated SuccessCode = "40001"
	SuccessVariantCreated SuccessCode = "40002"

	ErrProductNotFound ErrorCode = "40004"
	ErrVariantNotFound ErrorCode = "40004"
)

// ───────────────
// Transaction (Prefix 5): 50001++
// - Transaction management
// ───────────────

const (
	SuccessTransactionCreated        SuccessCode = "50001"
	SuccessTransactionDetailsCreated SuccessCode = "50002"

	ErrTransactionNotFound        ErrorCode = "50003"
	ErrTransactionDetailsNotFound ErrorCode = "50004"
)

// ───────────────
// Import (Prefix 6): 60001++
// - Import management
// ───────────────

const (
	SuccessImportCustomers    SuccessCode = "60001"
	SuccessImportTransactions SuccessCode = "60002"
	SuccessImportBatch        SuccessCode = "60003"

	ErrInvalidFilename         ErrorCode = "60004"
	ErrCustomerImportRequired  ErrorCode = "60005"
	ErrInvalidExcelFormat      ErrorCode = "60006"
	ErrInvalidBatchDate        ErrorCode = "60007"
	ErrBothFilesRequired       ErrorCode = "60008"
	ErrBatchProcessing         ErrorCode = "60009"
	ErrTransactionFileRequired ErrorCode = "60010"
)
