package parser

import (
	"fmt"
	"mime/multipart"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/xuri/excelize/v2"
)

// ReadCustomerExcel reads customer Excel file and returns slice of CustomerExcelRow
func ReadCustomerExcel(file multipart.File) ([]*model.CustomerExcelRow, error) {
	rows, err := readExcelRows(file)
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("excel file is empty or has no data rows")
	}

	return parseCustomerRows(rows), nil
}

// ReadTransactionExcel reads transaction Excel file and returns slice of TransactionExcelRow
func ReadTransactionExcel(file multipart.File) ([]*model.TransactionExcelRow, error) {
	rows, err := readExcelRows(file)
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("excel file is empty or has no data rows")
	}

	return parseTransactionRows(rows), nil
}

// readExcelRows reads all rows from Excel file
func readExcelRows(file multipart.File) ([][]string, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("no sheet found in Excel file")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	return rows, nil
}

// parseCustomerRows parses customer rows (skip header at index 0)
func parseCustomerRows(rows [][]string) []*model.CustomerExcelRow {
	var customers []*model.CustomerExcelRow

	for i := 1; i < len(rows); i++ {
		row := rows[i]

		if len(row) < 3 {
			continue
		}

		customer := &model.CustomerExcelRow{
			No:            i,
			NamaPelanggan: getStringValue(row, 1),
			NomorTelepon:  getStringValue(row, 2),
		}

		customers = append(customers, customer)
	}

	return customers
}

// parseTransactionRows parses transaction rows (skip header at index 0)
func parseTransactionRows(rows [][]string) []*model.TransactionExcelRow {
	var transactions []*model.TransactionExcelRow

	for i := 1; i < len(rows); i++ {
		row := rows[i]

		if len(row) < 21 {
			continue
		}

		transaction := &model.TransactionExcelRow{
			No:               i,
			NoStruk:          getStringValue(row, 1),
			Tanggal:          getStringValue(row, 2),
			Jam:              getStringValue(row, 3),
			NamaPelanggan:    getStringValue(row, 6),
			NamaProduk:       getStringValue(row, 7),
			JumlahProduk:     getIntValue(row, 8),
			HargaPerProduk:   getInt64Value(row, 9),
			Varian:           getStringValue(row, 10),
			HargaVarian:      getInt64Value(row, 11),
			Subtotal:         getInt64Value(row, 13),
			DiskonTransaksi:  getInt64Value(row, 14),
			OngkosKirim:      getInt64Value(row, 16),
			Total:            getInt64Value(row, 18),
			Status:           getStringValue(row, 19),
			MetodePembayaran: getStringValue(row, 20),
		}

		transactions = append(transactions, transaction)
	}

	return transactions
}

// getStringValue safely gets string value from row
func getStringValue(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return row[index]
}

// getIntValue safely gets int value from row
func getIntValue(row []string, index int) int {
	if index >= len(row) {
		return 0
	}
	var val int
	fmt.Sscanf(row[index], "%d", &val)
	return val
}

// getInt64Value safely gets int64 value from row
func getInt64Value(row []string, index int) int64 {
	if index >= len(row) {
		return 0
	}
	var val int64
	fmt.Sscanf(row[index], "%d", &val)
	return val
}
