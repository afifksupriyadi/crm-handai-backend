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
	})

	registerSuccesses([]detail{
		// General
		{string(SuccessOK), "Permintaan berhasil diproses", http.StatusOK},

		// Auth
		{string(SuccessLogin), "Berhasil login", http.StatusOK},

		// User
		{string(SuccessUserUpdated), "Berhasil memperbarui kata sandi", http.StatusOK},

		// Customer
		{string(SuccessCustomerCreated), "Customer berhasil dibuat", http.StatusCreated},
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
