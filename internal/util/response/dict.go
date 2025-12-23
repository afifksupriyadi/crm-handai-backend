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
		{string(ErrEmptyName), "Nama wajib diisi", http.StatusBadRequest},
		{string(ErrNameTooLong), "Nama terlalu panjang (maksimal 50 karakter)", http.StatusBadRequest},
		{string(ErrEmptyPhone), "Nomor telepon wajib diisi", http.StatusBadRequest},
		{string(ErrInvalidPhoneFormat), "Format nomor telepon tidak valid", http.StatusBadRequest},
		{string(ErrPhoneTooLong), "Nomor telepon terlalu panjang (maksimal 20 karakter)", http.StatusBadRequest},
		{string(ErrPhoneAlreadyExists), "Nomor telepon sudah terdaftar", http.StatusBadRequest},
		{string(ErrTransactionFileRequired), "File transaksi wajib diupload", http.StatusBadRequest},

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
		{string(ErrBothFilesRequired), "File customer dan transaction wajib diupload", http.StatusBadRequest},
		{string(ErrBatchProcessing), "Gagal memproses batch import", http.StatusInternalServerError},
		{string(ErrInvalidFilenameFormat), "Format nama file tidak valid. Expected: Transaksi_Pelanggan_Kasir_Warung_DDMMYY_HHMMSS.xlsx atau Transaksi_Kasir_Warung_DDMMYY_HHMMSS.xlsx", http.StatusBadRequest},
		{string(ErrTransactionDateExceedsCustomer), "Tanggal transaksi tidak boleh melebihi tanggal customer", http.StatusBadRequest},
		{string(ErrStartDateRequired), "start_date wajib diisi", http.StatusBadRequest},
		{string(ErrEndDateRequired), "end_date wajib diisi", http.StatusBadRequest},
		{string(ErrImportSequenceGap), "Terdapat gap dalam urutan import. Harap import data dari %s hingga %s terlebih dahulu", http.StatusBadRequest},

		// Analytics (Prefix 7)
		{string(ErrInvalidDateRange), "Rentang tanggal tidak valid", http.StatusBadRequest},
		{string(ErrInvalidPeriodType), "Tipe periode tidak valid (DAILY, MONTHLY, YEARLY)", http.StatusBadRequest},
		{string(ErrDateRangeTooLarge), "Rentang tanggal terlalu besar (maksimal 365 hari untuk DAILY)", http.StatusBadRequest},
		{string(ErrInvalidPreset), "Preset tidak valid", http.StatusBadRequest},
		{string(ErrInvalidDateFormat), "Format tanggal tidak valid (harus YYYY-MM-DD)", http.StatusBadRequest},
	})

	registerSuccesses([]detail{
		// General (Prefix 0)
		{string(SuccessOK), "Permintaan berhasil diproses", http.StatusOK},
		{string(SuccessWelcome), "Selamat datang di API CRM Handai", http.StatusOK},
		{string(SuccessHealthCheck), "API berjalan dengan baik", http.StatusOK},

		// Auth (Prefix 1)
		{string(SuccessLogin), "Berhasil login", http.StatusOK},

		// User (Prefix 2)
		{string(SuccessUserUpdated), "Berhasil memperbarui kata sandi", http.StatusOK},

		// Customer (Prefix 3)
		{string(SuccessCustomerCreated), "Customer berhasil dibuat", http.StatusCreated},
		{string(SuccessCustomerRetrieved), "Data customer berhasil diambil", http.StatusOK},
		{string(SuccessCustomerUpdated), "Customer berhasil diperbarui", http.StatusOK},
		{string(SuccessCustomerDeleted), "Customer berhasil dihapus", http.StatusOK},

		// Product (Prefix 4)
		{string(SuccessProductCreated), "Product berhasil dibuat", http.StatusCreated},
		{string(SuccessVariantCreated), "Variant berhasil dibuat", http.StatusCreated},

		// Transaction (Prefix 5)
		{string(SuccessTransactionCreated), "Transaction berhasil dibuat", http.StatusCreated},
		{string(SuccessTransactionDetailsCreated), "Transaction details berhasil dibuat", http.StatusCreated},

		// Import (Prefix 6)
		{string(SuccessImportCustomers), "Berhasil import data customer", http.StatusOK},
		{string(SuccessImportTransactions), "Berhasil import data transaksi", http.StatusOK},
		{string(SuccessImportBatch), "Berhasil import batch data", http.StatusOK},

		// Analytics (Prefix 7)
		{string(SuccessSalesChart), "Berhasil mengambil data sales chart", http.StatusOK},
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
