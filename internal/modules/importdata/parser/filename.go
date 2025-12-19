package parser

import (
	"fmt"
	"regexp"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/constant"
)

var (
	// Transaksi_Pelanggan_Kasir_Warung_201025_213949.xlsx
	customerFilenameRegex = regexp.MustCompile(`^Transaksi_Pelanggan_Kasir_Warung_(\d{6})_\d{6}\.xlsx$`)

	// Transaksi_Kasir_Warung_161025_143434.xlsx
	transactionFilenameRegex = regexp.MustCompile(`^Transaksi_Kasir_Warung_(\d{6})_\d{6}\.xlsx$`)
)

// ExtractBatchDateFromFilename extracts date from filename
// Format: DDMMYY (e.g., 161025 = 16 Oct 2025)
func ExtractBatchDateFromFilename(filename string, fileType string) (time.Time, error) {
	var matches []string

	if fileType == string(constant.ImportTypeCustomer) {
		matches = customerFilenameRegex.FindStringSubmatch(filename)
		if len(matches) < 2 {
			return time.Time{}, fmt.Errorf(
				"invalid customer filename format. Expected: Transaksi_Pelanggan_Kasir_Warung_DDMMYY_HHMMSS.xlsx, got: %s",
				filename,
			)
		}
	} else if fileType == string(constant.ImportTypeTransaction) {
		matches = transactionFilenameRegex.FindStringSubmatch(filename)
		if len(matches) < 2 {
			return time.Time{}, fmt.Errorf(
				"invalid transaction filename format. Expected: Transaksi_Kasir_Warung_DDMMYY_HHMMSS.xlsx, got: %s",
				filename,
			)
		}
	} else {
		return time.Time{}, fmt.Errorf("unknown file type: %s", fileType)
	}

	// Parse DDMMYY format
	dateStr := matches[1]                           // e.g., "161025"
	batchDate, err := time.Parse("020106", dateStr) // DD MM YY
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format in filename: %w", err)
	}

	return batchDate, nil
}
