package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"regexp"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/constant"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/parser"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"

	customerSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	customerModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"

	productSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products"
	productModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"

	transactionSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions"
	transactionModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
)

var (
	filenameDateRegex = regexp.MustCompile(`(\d{6})_\d+\.xlsx$`)
)

type ImportServiceImpl struct {
	customerService          customerSvc.CustomerService
	productService           productSvc.ProductService
	variantService           productSvc.VariantService
	transactionService       transactionSvc.TransactionService
	transactionDetailService transactionSvc.TransactionDetailService
	importLogRepo            repository.ImportLogRepository
}

func NewImportService(
	customerService customerSvc.CustomerService,
	productService productSvc.ProductService,
	variantService productSvc.VariantService,
	transactionService transactionSvc.TransactionService,
	transactionDetailService transactionSvc.TransactionDetailService,
	importLogRepo repository.ImportLogRepository,
) importdata.ImportService {
	return &ImportServiceImpl{
		customerService:          customerService,
		productService:           productService,
		variantService:           variantService,
		transactionService:       transactionService,
		transactionDetailService: transactionDetailService,
		importLogRepo:            importLogRepo,
	}
}

// ImportCustomers imports customer data from Excel file
func (s *ImportServiceImpl) ImportCustomers(ctx context.Context, file multipart.File, filename string) (*model.ImportCustomerResponse, error) {
	// Extract file date
	fileDate, err := extractDateFromFilename(filename)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidFilename, "Cannot parse date from filename")
	}

	// Read Excel
	rows, err := parser.ReadCustomerExcel(file)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to read Excel file")
	}

	// Process rows
	var (
		totalRows        = len(rows)
		successRows      = 0
		failedRows       = 0
		customersCreated = 0
		errors           []model.ImportRowError
	)

	for _, row := range rows {
		isNew, err := s.processCustomerRow(ctx, row)
		if err != nil {
			failedRows++
			errors = append(errors, model.ImportRowError{
				RowNumber: row.No + 1,
				Message:   err.Error(),
			})
			continue
		}

		successRows++
		if isNew {
			customersCreated++
		}
	}
	// Create import log
	if err := s.createImportLog(ctx, constant.ImportTypeCustomer, fileDate, filename, successRows); err != nil {
		// Log error but don't fail the import
		fmt.Printf("Failed to create import log: %v\n", err)
	}

	return &model.ImportCustomerResponse{
		TotalRows:        totalRows,
		SuccessRows:      successRows,
		FailedRows:       failedRows,
		CustomersCreated: customersCreated,
		Errors:           errors,
	}, nil
}

// ImportTransactions imports transaction data from Excel file
func (s *ImportServiceImpl) ImportTransactions(ctx context.Context, file multipart.File, filename string) (*model.ImportTransactionResponse, error) {
	// Extract file date
	fileDate, err := extractDateFromFilename(filename)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidFilename, "Cannot parse date from filename")
	}

	// Validate import order (customer must be imported first)
	if err := s.validateTransactionImportOrder(ctx, fileDate); err != nil {
		return nil, err
	}

	// Read Excel
	rows, err := parser.ReadTransactionExcel(file)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to read Excel file")
	}

	// Process rows
	var (
		totalRows              = len(rows)
		successRows            = 0
		failedRows             = 0
		transactionsWithMember = 0
		anonymousTransactions  = 0
		transactionsCreated    = make(map[string]bool) // Track unique transactions
		detailsCreated         = 0
		productsCreated        = 0
		variantsCreated        = 0
		errors                 []model.ImportRowError
	)

	for _, row := range rows {
		isNewTransaction, customerID, err := s.processTransactionRow(ctx, row)
		if err != nil {
			failedRows++
			errors = append(errors, model.ImportRowError{
				RowNumber: row.No + 1,
				Message:   err.Error(),
			})
			continue
		}

		successRows++
		detailsCreated++

		// Track unique transactions
		if isNewTransaction {
			transactionsCreated[row.NoStruk] = true
		}

		// Track member vs anonymous
		if customerID != nil {
			transactionsWithMember++
		} else {
			anonymousTransactions++
		}
	}

	// Create import log
	if err := s.createImportLog(ctx, constant.ImportTypeTransaction, fileDate, filename, successRows); err != nil {
		fmt.Printf("Failed to create import log: %v\n", err)
	}

	return &model.ImportTransactionResponse{
		TotalRows:   totalRows,
		SuccessRows: successRows,
		FailedRows:  failedRows,
		Summary: &model.TransactionImportSummary{
			TransactionsCreated:       len(transactionsCreated),
			TransactionsWithMember:    transactionsWithMember,
			AnonymousTransactions:     anonymousTransactions,
			TransactionDetailsCreated: detailsCreated,
			ProductsCreated:           productsCreated,
			VariantsCreated:           variantsCreated,
		},
		Errors: errors,
	}, nil
}

// processCustomerRow processes one customer row
func (s *ImportServiceImpl) processCustomerRow(ctx context.Context, row *model.CustomerExcelRow) (bool, error) {
	// Normalize customer data
	parsed, err := parser.NormalizeCustomer(row.NamaPelanggan, row.NomorTelepon)
	if err != nil {
		return false, fmt.Errorf("failed to normalize customer: %w", err)
	}

	// Check if customer already exists
	_, err = s.customerService.GetCustomerByPhone(ctx, parsed.Phone)
	isNew := err != nil

	// Get or create customer
	customer := &customerModel.Customer{
		Name:  parsed.Name,
		Phone: parsed.Phone,
	}

	_, err = s.customerService.GetOrCreateCustomer(ctx, customer)
	if err != nil {
		return false, fmt.Errorf("failed to create customer: %w", err)
	}

	return isNew, nil
}

// processTransactionRow processes one transaction row
// Returns: isNewTransaction, customerID, error
func (s *ImportServiceImpl) processTransactionRow(ctx context.Context, row *model.TransactionExcelRow) (bool, *int, error) {
	// 1. Try to find customer by name (normalized match)
	var customerID *int
	normalizedName := parser.NormalizeName(row.NamaPelanggan)
	customer, err := s.customerService.GetCustomerByName(ctx, normalizedName)
	if err == nil {
		customerID = &customer.ID
	}

	// 2. Parse and get/create product
	parsedProduct, err := parser.ParseProduct(row.NamaProduk)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse product: %w", err)
	}

	product := &productModel.Product{
		Name:      parsedProduct.NormalizedName,
		Category:  parsedProduct.Category,
		BasePrice: parsedProduct.BasePrice,
	}

	product, err = s.productService.GetOrCreateProduct(ctx, product)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get/create product: %w", err)
	}

	// 3. Parse and get/create variant (if applicable)
	var variantID *int
	parsedVariant, err := parser.ParseVariant(row.NamaProduk, row.Varian, row.HargaVarian, row.JumlahProduk)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse variant: %w", err)
	}

	if parsedProduct.HasVariants {
		variant := &productModel.Variant{
			ProductID:     product.ID,
			Name:          parsedVariant.Size,
			PriceModifier: parsedVariant.PriceModifier,
			IsDefault:     parsedVariant.IsDefault,
		}

		variant, err = s.variantService.GetOrCreateVariant(ctx, variant)
		if err != nil {
			return false, nil, fmt.Errorf("failed to get/create variant: %w", err)
		}
		variantID = &variant.ID
	}

	// 4. Parse transaction date
	transactionDate, err := parseTransactionDate(row.Tanggal, row.Jam)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse transaction date: %w", err)
	}

	// 5. Check if transaction exists, create if new
	_, err = s.transactionService.GetTransactionByCode(ctx, row.NoStruk)
	isNewTransaction := err != nil // Error means not found = new transaction

	if isNewTransaction {
		transaction := &transactionModel.Transaction{
			Code:            row.NoStruk,
			CustomerID:      customerID,
			TransactionDate: transactionDate,
			Discount:        row.DiskonTransaksi,
			ShippingCost:    row.OngkosKirim,
			PaymentMethod:   row.MetodePembayaran,
			Status:          row.Status,
		}

		_, err = s.transactionService.GetOrCreateTransaction(ctx, transaction)
		if err != nil {
			return false, nil, fmt.Errorf("failed to create transaction: %w", err)
		}
	}
	// If transaction exists, skip creation

	// 6. Calculate unit price and subtotal
	unitPrice := product.BasePrice + parsedVariant.PriceModifier
	subtotal := unitPrice * int64(row.JumlahProduk)

	// Validate subtotal matches Excel
	if subtotal != row.Subtotal {
		return false, nil, fmt.Errorf("subtotal mismatch: calculated %d, got %d", subtotal, row.Subtotal)
	}

	// 7. Create transaction detail
	detail := &transactionModel.TransactionDetail{
		TransactionCode: row.NoStruk,
		ProductID:       product.ID,
		VariantID:       variantID,
		Quantity:        row.JumlahProduk,
		UnitPrice:       unitPrice,
		Subtotal:        subtotal,
	}

	err = s.transactionDetailService.CreateTransactionDetail(ctx, detail)
	if err != nil {
		return false, nil, fmt.Errorf("failed to create transaction detail: %w", err)
	}

	return isNewTransaction, customerID, nil
}

// validateTransactionImportOrder validates that customer import exists for the date
func (s *ImportServiceImpl) validateTransactionImportOrder(ctx context.Context, fileDate time.Time) error {
	hasCustomerImport, err := s.importLogRepo.HasCustomerImportSinceDate(ctx, fileDate)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check customer import status")
	}

	if !hasCustomerImport {
		return response.WrapAppError(
			ctx,
			nil,
			response.ErrCustomerImportRequired,
			fmt.Sprintf("Customer import for date %s or later is required before importing transactions", fileDate.Format("2006-01-02")),
		)
	}

	return nil
}

// createImportLog creates import log entry
func (s *ImportServiceImpl) createImportLog(ctx context.Context, importType constant.ImportType, fileDate time.Time, filename string, rowsImported int) error {
	log := &model.ImportLog{
		ImportType:   importType.String(),
		FileDate:     fileDate,
		Filename:     filename,
		RowsImported: rowsImported,
		Status:       constant.ImportStatusSuccess.String(),
	}

	return s.importLogRepo.CreateImportLog(ctx, log)
}

// extractDateFromFilename extracts date from filename
// Format: Transaksi_Kasir_Warung_DDMMYY_HHMMSS.xlsx
func extractDateFromFilename(filename string) (time.Time, error) {
	matches := filenameDateRegex.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("filename does not match expected pattern")
	}

	dateStr := matches[1] // DDMMYY
	parsedDate, err := time.Parse("020106", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date from filename: %w", err)
	}

	return parsedDate, nil
}

// parseTransactionDate combines Tanggal and Jam into time.Time
// Format: Tanggal "01-09-2025", Jam "19:09:52"
func parseTransactionDate(tanggal, jam string) (time.Time, error) {
	datetime := fmt.Sprintf("%s %s", tanggal, jam)
	parsedTime, err := time.Parse("02-01-2006 15:04:05", datetime)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse transaction date: %w", err)
	}

	return parsedTime, nil
}
