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
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"

	customerSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	productSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products"
	productModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	transactionSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions"
	transactionModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
)

var (
	filenameDateRegex = regexp.MustCompile(`(\d{6})_\d+\.xlsx$`)
)

type ImportServiceImpl struct {
	db                       *bun.DB
	customerService          customerSvc.CustomerService
	productService           productSvc.ProductService
	variantService           productSvc.VariantService
	transactionService       transactionSvc.TransactionService
	transactionDetailService transactionSvc.TransactionDetailService
	importLogRepo            repository.ImportLogRepository
	batchRepo                repository.BatchRepository
}

func NewImportService(
	db *bun.DB,
	customerService customerSvc.CustomerService,
	productService productSvc.ProductService,
	variantService productSvc.VariantService,
	transactionService transactionSvc.TransactionService,
	transactionDetailService transactionSvc.TransactionDetailService,
	importLogRepo repository.ImportLogRepository,
	batchRepo repository.BatchRepository,
) importdata.ImportService {
	return &ImportServiceImpl{
		db:                       db,
		customerService:          customerService,
		productService:           productService,
		variantService:           variantService,
		transactionService:       transactionService,
		transactionDetailService: transactionDetailService,
		importLogRepo:            importLogRepo,
		batchRepo:                batchRepo,
	}
}

// ==========================================
// MAIN BATCH IMPORT FUNCTION
// ==========================================

// ImportBatch imports transaction file (required) and optionally customer file
func (s *ImportServiceImpl) ImportBatch(
	ctx context.Context,
	customerFile, transactionFile multipart.File,
	customerFilename, transactionFilename,
	batchDateStr, notes string,
) (*model.ImportBatchResponse, error) {
	// 1. Validate batch date
	batchDate, err := time.Parse("2006-01-02", batchDateStr)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidBatchDate, "Invalid batch date format (expected YYYY-MM-DD)")
	}

	// 2. Start database transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start transaction")
	}
	defer tx.Rollback()

	logger.Get().Info().
		Str("batch_date", batchDateStr).
		Bool("has_customer_file", customerFile != nil).
		Msg("Starting batch import with transaction")

	// 3. Create batch entry with PROCESSING status
	batch := &model.Batch{
		BatchDate: batchDate,
		BatchCode: generateBatchCode(batchDate),
		Status:    "PROCESSING",
		IsActive:  false,
	}

	if err := s.batchRepo.CreateBatch(ctx, &tx, batch); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to create batch")
	}

	logger.Get().Info().
		Int("batch_id", batch.ID).
		Str("batch_code", batch.BatchCode).
		Msg("Batch created with PROCESSING status")

	// 4. Import customers (OPTIONAL)
	var customerSummary *model.ImportCustomerSummary
	var customerImportLog *model.ImportLog

	if customerFile != nil {
		logger.Get().Info().Msg("Customer file provided, importing customers first")
		customerSummary, customerImportLog, err = s.importCustomersInBatch(ctx, &tx, batch.ID, customerFile, customerFilename, batchDate)
		if err != nil {
			s.batchRepo.UpdateBatchStatus(ctx, &tx, batch.ID, "FAILED")
			return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to import customers")
		}
	} else {
		logger.Get().Info().Msg("No customer file provided, customers will be auto-created from transactions")
		// Create dummy summary
		customerSummary = &model.ImportCustomerSummary{
			TotalRows:        0,
			SuccessRows:      0,
			FailedRows:       0,
			CustomersCreated: 0,
			CustomersUpdated: 0,
			Errors:           []model.ImportRowError{},
		}
	}

	// 5. Import transactions (REQUIRED)
	transactionSummary, transactionImportLog, err := s.importTransactionsInBatch(ctx, &tx, batch.ID, transactionFile, transactionFilename, batchDate)
	if err != nil {
		s.batchRepo.UpdateBatchStatus(ctx, &tx, batch.ID, "FAILED")
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to import transactions")
	}

	// 6. Update customer metrics for all affected customers
	if err := s.updateAllCustomerMetrics(ctx, &tx); err != nil {
		logger.Get().Warn().Err(err).Msg("Failed to update customer metrics, continuing anyway")
	}

	// 7. Link import logs to batch
	var customerImportID *int
	if customerImportLog != nil {
		customerImportID = &customerImportLog.ID
	}
	transactionImportID := transactionImportLog.ID

	if err := s.batchRepo.LinkImportLogs(ctx, &tx, batch.ID, customerImportID, &transactionImportID); err != nil {
		s.batchRepo.UpdateBatchStatus(ctx, &tx, batch.ID, "FAILED")
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to link import logs")
	}

	// 8. Update batch status to COMPLETED and set as active
	if err := s.batchRepo.UpdateBatchStatus(ctx, &tx, batch.ID, "COMPLETED"); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to update batch status")
	}

	if err := s.batchRepo.SetActiveBatch(ctx, &tx, batch.ID); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to set active batch")
	}

	// 9. Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit transaction")
	}

	logger.Get().Info().
		Int("batch_id", batch.ID).
		Str("status", "COMPLETED").
		Int("customers_from_file", customerSummary.CustomersCreated).
		Int("customers_from_transactions", transactionSummary.TransactionsCreated).
		Msg("Batch import completed successfully")

	// 10. Build response
	return &model.ImportBatchResponse{
		Batch: &model.BatchInfo{
			ID:        batch.ID,
			BatchCode: batch.BatchCode,
			BatchDate: batch.BatchDate.Format("2006-01-02"),
			Status:    "COMPLETED",
			IsActive:  true,
		},
		Customers:    customerSummary,
		Transactions: transactionSummary,
	}, nil
}

// ==========================================
// CUSTOMER IMPORT (Within Transaction)
// ==========================================

func (s *ImportServiceImpl) importCustomersInBatch(
	ctx context.Context,
	tx *bun.Tx,
	_ int, // batchID - reserved for future use
	file multipart.File,
	filename string,
	batchDate time.Time,
) (*model.ImportCustomerSummary, *model.ImportLog, error) {
	// Read Excel
	rows, err := parser.ReadCustomerExcel(file)
	if err != nil {
		return nil, nil, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to read customer Excel file")
	}

	logger.Get().Info().
		Int("total_rows", len(rows)).
		Msg("Starting customer import")

	// Process rows
	var (
		totalRows        = len(rows)
		successRows      = 0
		failedRows       = 0
		customersCreated = 0
		customersUpdated = 0
		errors           []model.ImportRowError
	)

	for _, row := range rows {
		isNew, err := s.processCustomerRowInBatch(ctx, tx, row)
		if err != nil {
			failedRows++
			errors = append(errors, model.ImportRowError{
				RowNumber: row.No + 1,
				Message:   err.Error(),
			})
			logger.Get().Error().
				Int("row", row.No+1).
				Err(err).
				Msg("Failed to process customer row")
			continue
		}

		successRows++
		if isNew {
			customersCreated++
		} else {
			customersUpdated++
		}
	}

	// Create import log
	importLog := &model.ImportLog{
		ImportType:   constant.ImportTypeCustomer.String(),
		FileDate:     batchDate,
		Filename:     filename,
		RowsImported: successRows,
		Status:       constant.ImportStatusSuccess.String(),
	}

	if err := s.createImportLogInBatch(ctx, tx, importLog); err != nil {
		return nil, nil, err
	}

	logger.Get().Info().
		Int("total_rows", totalRows).
		Int("success_rows", successRows).
		Int("failed_rows", failedRows).
		Int("created", customersCreated).
		Int("updated", customersUpdated).
		Msg("Customer import completed")

	summary := &model.ImportCustomerSummary{
		TotalRows:        totalRows,
		SuccessRows:      successRows,
		FailedRows:       failedRows,
		CustomersCreated: customersCreated,
		CustomersUpdated: customersUpdated,
		Errors:           errors,
	}

	return summary, importLog, nil
}

func (s *ImportServiceImpl) processCustomerRowInBatch(ctx context.Context, tx *bun.Tx, row *model.CustomerExcelRow) (bool, error) {
	// Normalize customer data
	parsed, err := parser.NormalizeCustomer(row.NamaPelanggan, row.NomorTelepon)
	if err != nil {
		return false, fmt.Errorf("failed to normalize customer: %w", err)
	}

	// Find or create customer with name matching logic
	_, isNew, err := s.customerService.FindOrCreateCustomerWithNameMatching(ctx, tx, parsed.Name, parsed.Phone)
	if err != nil {
		return false, fmt.Errorf("failed to find/create customer: %w", err)
	}

	return isNew, nil
}

// ==========================================
// TRANSACTION IMPORT (Within Transaction)
// ==========================================

func (s *ImportServiceImpl) importTransactionsInBatch(
	ctx context.Context,
	tx *bun.Tx,
	batchID int,
	file multipart.File,
	filename string,
	batchDate time.Time,
) (*model.ImportTransactionSummary, *model.ImportLog, error) {
	// Read Excel
	rows, err := parser.ReadTransactionExcel(file)
	if err != nil {
		return nil, nil, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to read transaction Excel file")
	}

	logger.Get().Info().
		Int("total_rows", len(rows)).
		Msg("Starting transaction import")

	// Process rows
	var (
		totalRows           = len(rows)
		successRows         = 0
		failedRows          = 0
		transactionsCreated = make(map[string]bool) // Track unique transactions
		detailsCreated      = 0
		productsCreated     = 0
		variantsCreated     = 0
		errors              []model.ImportRowError
		batchSize           = 100 // Commit every 100 rows to prevent connection timeout
	)

	for i, row := range rows {
		// Commit and restart transaction every batch to prevent timeout
		if i%batchSize == 0 && i > 0 {
			// Commit current batch
			if err := tx.Commit(); err != nil {
				logger.Get().Error().Err(err).Int("row", i).Msg("Failed to commit batch")
				return nil, nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit batch")
			}

			logger.Get().Info().Int("processed", i).Int("total", totalRows).Msg("Batch committed, starting new transaction")

			// Start new transaction for next batch
			newTx, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				logger.Get().Error().Err(err).Int("row", i).Msg("Failed to start new transaction")
				return nil, nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start new transaction")
			}
			tx = &newTx
		}

		isNewTransaction, err := s.processTransactionRowInBatch(ctx, tx, batchID, row)
		if err != nil {
			failedRows++
			errors = append(errors, model.ImportRowError{
				RowNumber: row.No + 1,
				Message:   err.Error(),
			})
			logger.Get().Error().
				Int("row", row.No+1).
				Err(err).
				Msg("Failed to process transaction row")
			continue
		}

		successRows++
		detailsCreated++

		// Track unique transactions
		if isNewTransaction {
			transactionsCreated[row.NoStruk] = true
		}
	}

	// Create import log
	importLog := &model.ImportLog{
		ImportType:   constant.ImportTypeTransaction.String(),
		FileDate:     batchDate,
		Filename:     filename,
		RowsImported: successRows,
		Status:       constant.ImportStatusSuccess.String(),
	}

	if err := s.createImportLogInBatch(ctx, tx, importLog); err != nil {
		return nil, nil, err
	}

	logger.Get().Info().
		Int("total_rows", totalRows).
		Int("success_rows", successRows).
		Int("failed_rows", failedRows).
		Int("transactions_created", len(transactionsCreated)).
		Int("details_created", detailsCreated).
		Msg("Transaction import completed")

	summary := &model.ImportTransactionSummary{
		TotalRows:                 totalRows,
		SuccessRows:               successRows,
		FailedRows:                failedRows,
		TransactionsCreated:       len(transactionsCreated),
		TransactionDetailsCreated: detailsCreated,
		ProductsCreated:           productsCreated,
		VariantsCreated:           variantsCreated,
		Errors:                    errors,
	}

	return summary, importLog, nil
}

func (s *ImportServiceImpl) processTransactionRowInBatch(
	ctx context.Context,
	tx *bun.Tx,
	batchID int,
	row *model.TransactionExcelRow,
) (bool, error) {
	// 1. Try to find customer by name (normalized match)
	var customerID *int
	normalizedName := parser.NormalizeName(row.NamaPelanggan)
	customer, err := s.customerService.GetCustomerByName(ctx, normalizedName)
	if err == nil && customer != nil {
		customerID = &customer.ID
	}

	// 2. Parse and get/create product
	parsedProduct, err := parser.ParseProduct(row.NamaProduk)
	if err != nil {
		return false, fmt.Errorf("failed to parse product: %w", err)
	}

	product := &productModel.Product{
		Name:      parsedProduct.NormalizedName,
		Category:  parsedProduct.Category,
		BasePrice: parsedProduct.BasePrice,
	}

	product, err = s.productService.GetOrCreateProductInTx(ctx, tx, product)
	if err != nil {
		return false, fmt.Errorf("failed to get/create product: %w", err)
	}

	// 3. Parse and get/create variant (if applicable)
	var variantID *int
	parsedVariant, err := parser.ParseVariant(row.NamaProduk, row.Varian, row.HargaVarian, row.JumlahProduk)
	if err != nil {
		return false, fmt.Errorf("failed to parse variant: %w", err)
	}

	if parsedProduct.HasVariants {
		variant := &productModel.Variant{
			ProductID:     product.ID,
			Name:          parsedVariant.Size,
			PriceModifier: parsedVariant.PriceModifier,
			IsDefault:     parsedVariant.IsDefault,
		}

		variant, err = s.variantService.GetOrCreateVariantInTx(ctx, tx, variant)
		if err != nil {
			return false, fmt.Errorf("failed to get/create variant: %w", err)
		}
		variantID = &variant.ID
	}

	// 4. Parse transaction date
	transactionDate, err := parseTransactionDate(row.Tanggal, row.Jam)
	if err != nil {
		return false, fmt.Errorf("failed to parse transaction date: %w", err)
	}

	// 5. Check if transaction exists, create if new
	existingTransaction, _ := s.transactionService.GetTransactionByCodeInTx(ctx, tx, row.NoStruk)
	isNewTransaction := (existingTransaction == nil)

	if isNewTransaction {
		transaction := &transactionModel.Transaction{
			Code:            row.NoStruk,
			CustomerID:      customerID,
			TransactionDate: transactionDate,
			Discount:        row.DiskonTransaksi,
			ShippingCost:    row.OngkosKirim,
			PaymentMethod:   row.MetodePembayaran,
			Status:          row.Status,
			BatchID:         &batchID, // Link to batch
		}

		if err := s.transactionService.CreateTransactionInTx(ctx, tx, transaction); err != nil {
			return false, fmt.Errorf("failed to create transaction: %w", err)
		}
	}

	// 6. Calculate unit price and subtotal
	unitPrice := product.BasePrice + parsedVariant.PriceModifier
	subtotal := unitPrice * float64(row.JumlahProduk)

	// Validate subtotal matches Excel
	if subtotal != row.Subtotal {
		return false, fmt.Errorf("subtotal mismatch: calculated %.2f, got %.2f", subtotal, row.Subtotal)
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

	if err := s.transactionDetailService.CreateTransactionDetailInTx(ctx, tx, detail); err != nil {
		return false, fmt.Errorf("failed to create transaction detail: %w", err)
	}

	return isNewTransaction, nil
}

// ==========================================
// CUSTOMER METRICS UPDATE
// ==========================================

func (s *ImportServiceImpl) updateAllCustomerMetrics(ctx context.Context, tx *bun.Tx) error {
	// Get all customers who have transactions
	query := `
		SELECT DISTINCT customer_id 
		FROM transactions 
		WHERE customer_id IS NOT NULL 
		AND deleted_at IS NULL
	`

	var customerIDs []int
	if err := (*tx).NewRaw(query).Scan(ctx, &customerIDs); err != nil {
		return fmt.Errorf("failed to get customer IDs: %w", err)
	}

	logger.Get().Info().
		Int("total_customers", len(customerIDs)).
		Msg("Updating customer metrics")

	// Update metrics for each customer
	for _, customerID := range customerIDs {
		if err := s.customerService.UpdateCustomerMetrics(ctx, tx, customerID); err != nil {
			logger.Get().Warn().
				Int("customer_id", customerID).
				Err(err).
				Msg("Failed to update customer metrics")
			// Don't fail the whole import, just log warning
		}
	}

	logger.Get().Info().Msg("Customer metrics updated")
	return nil
}

// ==========================================
// HELPER FUNCTIONS
// ==========================================

func (s *ImportServiceImpl) createImportLogInBatch(ctx context.Context, tx *bun.Tx, log *model.ImportLog) error {
	_, err := (*tx).NewInsert().
		Model(log).
		Exec(ctx)

	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create import log")
	}

	return nil
}

func generateBatchCode(batchDate time.Time) string {
	return fmt.Sprintf("BATCH_%s", batchDate.Format("20060102"))
}

func parseTransactionDate(tanggal, jam string) (time.Time, error) {
	datetime := fmt.Sprintf("%s %s", tanggal, jam)
	parsedTime, err := time.Parse("02-01-2006 15:04:05", datetime)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse transaction date: %w", err)
	}
	return parsedTime, nil
}

// ==========================================
// LEGACY ENDPOINTS (Backward Compatibility)
// ==========================================

// ImportCustomers - Legacy endpoint (keep for backward compatibility)
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

	// Process rows (without transaction)
	var (
		totalRows        = len(rows)
		successRows      = 0
		failedRows       = 0
		customersCreated = 0
		errors           []model.ImportRowError
	)

	for _, row := range rows {
		isNew, err := s.processCustomerRowLegacy(ctx, row)
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

func (s *ImportServiceImpl) processCustomerRowLegacy(ctx context.Context, row *model.CustomerExcelRow) (bool, error) {
	parsed, err := parser.NormalizeCustomer(row.NamaPelanggan, row.NomorTelepon)
	if err != nil {
		return false, fmt.Errorf("failed to normalize customer: %w", err)
	}

	existingCustomer, _ := s.customerService.GetCustomerByPhone(ctx, parsed.Phone)
	isNew := (existingCustomer == nil)

	_, err = s.customerService.GetOrCreateCustomer(ctx, parsed.Name, parsed.Phone)
	if err != nil {
		return false, fmt.Errorf("failed to create customer: %w", err)
	}

	return isNew, nil
}

// ImportTransactions - Legacy endpoint (keep for backward compatibility)
func (s *ImportServiceImpl) ImportTransactions(ctx context.Context, file multipart.File, filename string) (*model.ImportTransactionResponse, error) {
	// Extract file date
	fileDate, err := extractDateFromFilename(filename)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidFilename, "Cannot parse date from filename")
	}

	// Validate import order
	if err := s.validateTransactionImportOrder(ctx, fileDate); err != nil {
		return nil, err
	}

	// Read Excel
	rows, err := parser.ReadTransactionExcel(file)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to read Excel file")
	}

	// Process rows (without transaction)
	var (
		totalRows           = len(rows)
		successRows         = 0
		failedRows          = 0
		transactionsCreated = make(map[string]bool)
		detailsCreated      = 0
		errors              []model.ImportRowError
	)

	for _, row := range rows {
		isNewTransaction, _, err := s.processTransactionRowLegacy(ctx, row)
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

		if isNewTransaction {
			transactionsCreated[row.NoStruk] = true
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
			TransactionDetailsCreated: detailsCreated,
		},
		Errors: errors,
	}, nil
}

func (s *ImportServiceImpl) processTransactionRowLegacy(ctx context.Context, row *model.TransactionExcelRow) (bool, *int, error) {
	// Similar to batch processing but without transaction context
	// Implementation sama seperti yang sudah ada sebelumnya
	// (copy dari code yang sudah ada di import_service.go lama)

	var customerID *int
	normalizedName := parser.NormalizeName(row.NamaPelanggan)
	customer, err := s.customerService.GetCustomerByName(ctx, normalizedName)
	if err == nil && customer != nil {
		customerID = &customer.ID
	}

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

	transactionDate, err := parseTransactionDate(row.Tanggal, row.Jam)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse transaction date: %w", err)
	}

	_, err = s.transactionService.GetTransactionByCode(ctx, row.NoStruk)
	isNewTransaction := err != nil

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

	unitPrice := product.BasePrice + parsedVariant.PriceModifier
	subtotal := unitPrice * float64(row.JumlahProduk)

	if subtotal != row.Subtotal {
		return false, nil, fmt.Errorf("subtotal mismatch: calculated %.2f, got %.2f", subtotal, row.Subtotal)
	}

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

func extractDateFromFilename(filename string) (time.Time, error) {
	matches := filenameDateRegex.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("filename does not match expected pattern")
	}

	dateStr := matches[1]
	parsedDate, err := time.Parse("020106", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date from filename: %w", err)
	}

	return parsedDate, nil
}
