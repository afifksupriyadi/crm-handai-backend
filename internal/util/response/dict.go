package response

import (
	"fmt"
	"net/http"
)

// detail represents a response detail containing code, message, and HTTP status
type detail struct {
	// Code is the unique identifier for the response
	Code string `json:"code"`
	// Message is the human-readable response message
	Message string `json:"message"`
	// HTTPStatus is the corresponding HTTP status code
	HTTPStatus int
}

var (
	errorDict   = map[ErrorCode]detail{}
	successDict = map[SuccessCode]detail{}
)

// Register application custom codes and message here.
// You can add more codes as needed, following this format:
// {string(ErrorCode), "Message in Bahasa Indonesia", HTTPStatusCode} or
// {string(SuccessCode), "Message in Bahasa Indonesia", HTTPStatusCode}
func init() {
	registerErrors([]detail{
		// General (Prefix 0)
		{string(ErrInternalServerError), "Terjadi kesalahan pada server internal", http.StatusInternalServerError},
		{string(ErrEmptyRequestBody), "Request body wajib diisi", http.StatusBadRequest},
		{string(ErrDatabaseError), "Terjadi kesalahan pada database", http.StatusInternalServerError},
		{string(ErrUnprocessableEntity), "Terjadi kesalahan pada data yang dikirim", http.StatusBadRequest},

		// Auth (Prefix 1)
		{string(ErrEmptyEmail), "Email wajib diisi", http.StatusBadRequest},
		{string(ErrEmptyPassword), "Kata sandi wajib diisi", http.StatusBadRequest},
		{string(ErrInvalidEmailFormat), "Format email tidak valid", http.StatusBadRequest},
		{string(ErrInvalidCredentials), "Kata sandi salah. Silahkan coba lagi", http.StatusBadRequest},
		{string(ErrTokenNotFound), "Token tidak ditemukan", http.StatusUnauthorized},
		{string(ErrInvalidToken), "Token tidak valid", http.StatusUnauthorized},

		// User (Prefix 2)
		{string(ErrUserNotFound), "User tidak ditemukan", http.StatusNotFound},

		// Customer (Prefix 3)
		{string(ErrCustomerNotFound), "Customer tidak ditemukan", http.StatusNotFound},

		// Product (Prefix 4)
		{string(ErrProductNotFound), "Product tidak ditemukan", http.StatusNotFound},
		{string(ErrVariantNotFound), "Variant tidak ditemukan", http.StatusNotFound},

		// Transaction (Prefix 5)
		{string(ErrTransactionNotFound), "Transaction tidak ditemukan", http.StatusNotFound},
		{string(ErrTransactionDetailsNotFound), "Transaction details tidak ditemukan", http.StatusNotFound},

		// Import (Prefix 6)
		{string(ErrInvalidFilename), "Format nama file tidak valid", http.StatusBadRequest},
		{string(ErrCustomerImportRequired), "Import customer harus dilakukan terlebih dahulu", http.StatusBadRequest},
		{string(ErrInvalidExcelFormat), "Format file Excel tidak valid", http.StatusBadRequest},
	})

	registerSuccesses([]detail{
		// General (Prefix 0)
		{string(SuccessOK), "Permintaan berhasil diproses", http.StatusOK},

		// Auth (Prefix 1)
		{string(SuccessLogin), "Berhasil login", http.StatusOK},

		// User (Prefix 2)
		{string(SuccessUserUpdated), "Berhasil memperbarui kata sandi", http.StatusOK},

		// Customer (Prefix 3)
		{string(SuccessCustomerCreated), "Customer berhasil dibuat", http.StatusCreated},

		// Product (Prefix 4)
		{string(SuccessProductCreated), "Product berhasil dibuat", http.StatusCreated},
		{string(SuccessVariantCreated), "Variant berhasil dibuat", http.StatusCreated},

		// Transaction (Prefix 5)
		{string(SuccessTransactionCreated), "Transaction berhasil dibuat", http.StatusCreated},
		{string(SuccessTransactionDetailsCreated), "Transaction details berhasil dibuat", http.StatusCreated},

		// Import (Prefix 6)
		{string(SuccessImportCustomers), "Berhasil import data customer", http.StatusOK},
		{string(SuccessImportTransactions), "Berhasil import data transaksi", http.StatusOK},
	})
}

func registerErrors(list []detail) {
	for _, d := range list {
		errorDict[ErrorCode(d.Code)] = d
	}
}

func registerSuccesses(list []detail) {
	for _, d := range list {
		successDict[SuccessCode(d.Code)] = d
	}
}

func getDetail[T comparable](dict map[T]detail, code T, args ...interface{}) (*detail, bool) {
	d, ok := dict[code]
	if !ok {
		return nil, false
	}

	if len(args) > 0 {
		d.Message = fmt.Sprintf(d.Message, args...)
	}

	return &d, true
}

func getErrorDetail(code ErrorCode, args ...interface{}) (*detail, bool) {
	return getDetail(errorDict, code, args...)
}

func getSuccessDetail(code SuccessCode, args ...interface{}) (*detail, bool) {
	return getDetail(successDict, code, args...)
}
