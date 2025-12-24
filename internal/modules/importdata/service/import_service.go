package service

import (
	"context"
	"fmt"
	"mime/multipart"
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
	customerModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	forecastingSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting"
	productSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products"
	productModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	transactionSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions"
	transactionModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
)

// ImportServiceImpl implements ImportService interface
type ImportServiceImpl struct {
	db                       *bun.DB
	customerService          customerSvc.CustomerService
	productService           productSvc.ProductService
	variantService           productSvc.VariantService
	transactionService       transactionSvc.TransactionService
	transactionDetailService transactionSvc.TransactionDetailService
	customerBatchRepo        repository.CustomerBatchRepository
	transactionBatchRepo     repository.TransactionBatchRepository
	importLogRepo            repository.ImportLogRepository
	importTrackerRepo        repository.ImportTrackerRepository
	predictionOrchestrator   customerSvc.PredictionOrchestratorService
	forecastingService       forecastingSvc.SalesForecastService
}

// NewImportService creates a new instance of ImportServiceImpl
func NewImportService(
	db *bun.DB,
	customerService customerSvc.CustomerService,
	productService productSvc.ProductService,
	variantService productSvc.VariantService,
	transactionService transactionSvc.TransactionService,
	transactionDetailService transactionSvc.TransactionDetailService,
	customerBatchRepo repository.CustomerBatchRepository,
	transactionBatchRepo repository.TransactionBatchRepository,
	importLogRepo repository.ImportLogRepository,
	importTrackerRepo repository.ImportTrackerRepository,
	predictionOrchestrator customerSvc.PredictionOrchestratorService,
	forecastingService forecastingSvc.SalesForecastService,
) importdata.ImportService {
	return &ImportServiceImpl{
		db:                       db,
		customerService:          customerService,
		productService:           productService,
		variantService:           variantService,
		transactionService:       transactionService,
		transactionDetailService: transactionDetailService,
		customerBatchRepo:        customerBatchRepo,
		transactionBatchRepo:     transactionBatchRepo,
		importLogRepo:            importLogRepo,
		importTrackerRepo:        importTrackerRepo,
		predictionOrchestrator:   predictionOrchestrator,
		forecastingService:       forecastingService,
	}
}

// ImportBatch processes batch import with sequential validation and predictions
func (s *ImportServiceImpl) ImportBatch(ctx context.Context, customerFile, transactionFile multipart.File, customerFilename, transactionFilename, notes string, startDate, endDate time.Time) (*model.ImportBatchResponse, error) {
	log := logger.FromContext(ctx, 2)

	// ========== SEQUENTIAL VALIDATION ==========
	tracker, err := s.importTrackerRepo.GetLatest(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get import tracker")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to validate import sequence")
	}

	// Define default/initial date (from migration)
	defaultDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	// Check if this is first import (tracker still has default value)
	if tracker == nil || tracker.LastImportEndDate.Equal(defaultDate) {
		log.Info().Msg("First import detected, skipping sequential validation")
	} else {
		// This is NOT first import, validate sequence
		expectedStartDate := tracker.LastImportEndDate.AddDate(0, 0, 1)

		if !startDate.Equal(expectedStartDate) {
			missingStartDate := expectedStartDate.Format("2006-01-02")
			missingEndDate := startDate.AddDate(0, 0, -1).Format("2006-01-02")

			log.Error().
				Str("expected_start", missingStartDate).
				Str("got_start", startDate.Format("2006-01-02")).
				Str("last_import_end", tracker.LastImportEndDate.Format("2006-01-02")).
				Str("missing_range", fmt.Sprintf("%s to %s", missingStartDate, missingEndDate)).
				Msg("Import sequence gap detected")

			return nil, response.WrapAppError(
				ctx,
				fmt.Errorf("import sequence gap: expected %s, got %s", missingStartDate, startDate.Format("2006-01-02")),
				response.ErrImportSequenceGap,
				"Sequential import validation failed",
				missingStartDate,
				missingEndDate,
			)
		}

		log.Info().
			Str("last_import_end", tracker.LastImportEndDate.Format("2006-01-02")).
			Str("current_start", startDate.Format("2006-01-02")).
			Msg("Sequential validation passed")
	}

	// ========== VALIDATE FILENAME FORMAT ==========
	if err := parser.ValidateFilenameFormat(transactionFilename, string(constant.ImportTypeTransaction)); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidFilenameFormat, err.Error())
	}

	if customerFile != nil {
		if err := parser.ValidateFilenameFormat(customerFilename, string(constant.ImportTypeCustomer)); err != nil {
			return nil, response.WrapAppError(ctx, err, response.ErrInvalidFilenameFormat, err.Error())
		}
	}

	// ========== START MAIN TRANSACTION ==========
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start transaction")
	}
	defer tx.Rollback()

	log.Info().
		Str("start_date", startDate.Format("2006-01-02")).
		Str("end_date", endDate.Format("2006-01-02")).
		Bool("has_customer_file", customerFile != nil).
		Msg("Starting batch import")

	// ========== CREATE CUSTOMER BATCH ==========
	customerBatch := &model.CustomerBatch{
		BatchDate: endDate,
		Filename:  customerFilename,
		Notes:     notes,
		IsActive:  false,
	}
	createdCustomerBatch, err := s.customerBatchRepo.Create(ctx, &tx, customerBatch)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to create customer batch")
	}

	log.Info().Int("customer_batch_id", createdCustomerBatch.ID).Msg("Customer batch created")

	// ========== IMPORT CUSTOMERS ==========
	var customerSummary *model.ImportCustomerSummary
	var customerImportLog *model.ImportLog

	if customerFile != nil {
		log.Info().Msg("Importing customers from file")
		customerSummary, customerImportLog, err = s.importCustomersInBatch(ctx, &tx, createdCustomerBatch.ID, customerFile, customerFilename, endDate)
		if err != nil {
			return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to import customers")
		}
		log.Info().Int("customer_import_log_id", customerImportLog.ID).Msg("Customer import log created")

		createdCustomerBatch.CustomerCount = customerSummary.TotalRows
		createdCustomerBatch.NewCustomers = customerSummary.CustomersCreated
		createdCustomerBatch.UpdatedCustomers = customerSummary.CustomersUpdated
		_, err = s.customerBatchRepo.Update(ctx, &tx, createdCustomerBatch)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to update customer batch counts")
		}
	} else {
		log.Info().Msg("No customer file provided, transactions will save guests")
		customerSummary = &model.ImportCustomerSummary{
			TotalRows:        0,
			SuccessRows:      0,
			FailedRows:       0,
			CustomersCreated: 0,
			CustomersUpdated: 0,
			Errors:           []model.ImportRowError{},
		}
	}

	// ========== CREATE TRANSACTION BATCH ==========
	transactionBatch := &model.TransactionBatch{
		BatchDate:       endDate,
		Filename:        transactionFilename,
		CustomerBatchID: createdCustomerBatch.ID,
		Notes:           notes,
	}
	createdTransactionBatch, err := s.transactionBatchRepo.Create(ctx, &tx, transactionBatch)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to create transaction batch")
	}

	log.Info().Int("transaction_batch_id", createdTransactionBatch.ID).Msg("Transaction batch created")

	// ========== COMMIT MAIN TRANSACTION ==========
	if err := tx.Commit(); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit main transaction")
	}

	// ========== IMPORT TRANSACTIONS (BATCHED) ==========
	transactionSummary, transactionImportLog, _, err := s.importTransactionsInBatch(ctx, createdTransactionBatch.ID, transactionFile, transactionFilename, endDate)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to import transactions")
	}
	log.Info().Int("transaction_import_log_id", transactionImportLog.ID).Msg("Transaction import log created")

	// ========== FINAL TRANSACTION FOR UPDATES ==========
	finalTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start final transaction")
	}
	defer finalTx.Rollback()

	createdTransactionBatch.TransactionCount = transactionSummary.TransactionsCreated
	createdTransactionBatch.RegisteredTransactions = transactionSummary.TransactionsCreated - transactionSummary.FailedRows
	_, err = s.transactionBatchRepo.Update(ctx, &finalTx, createdTransactionBatch)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to update transaction batch counts")
	}

	err = s.customerBatchRepo.SetActive(ctx, &finalTx, createdCustomerBatch.ID)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to set active batch")
	}

	if err := finalTx.Commit(); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit final transaction")
	}

	log.Info().Int("customer_batch_id", createdCustomerBatch.ID).Int("transaction_batch_id", createdTransactionBatch.ID).Msg("Batch import committed, starting predictions")

	// ========== PROCESS PREDICTIONS ==========
	if err := s.predictionOrchestrator.ProcessPredictions(ctx, startDate, endDate, createdTransactionBatch.ID); err != nil {
		log.Warn().Err(err).Msg("Failed to process predictions (non-critical)")
	} else {
		log.Info().Msg("Predictions processed successfully")
	}

	// ========== GENERATE FORECASTS ========== ✅ ADD THIS BLOCK
	if s.forecastingService != nil {
		log.Info().Msg("Starting forecast generation")
		forecastResp, err := s.forecastingService.GenerateForecasts(ctx, endDate, createdTransactionBatch.ID)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to generate forecasts (non-critical)")
		} else {
			log.Info().
				Int("weekly", forecastResp.WeeklyForecasts).
				Int("monthly", forecastResp.MonthlyForecasts).
				Int("yearly", forecastResp.YearlyForecasts).
				Msg("Forecasts generated successfully")
		}
	}

	log.Info().Msg("Batch import completed successfully")

	return &model.ImportBatchResponse{
		Batch: &model.BatchInfo{
			ID:        createdTransactionBatch.ID,
			BatchCode: fmt.Sprintf("TB_%s", endDate.Format("20060102")),
			BatchDate: endDate.Format("2006-01-02"),
			Status:    "COMPLETED",
			IsActive:  true,
		},
		Customers:    customerSummary,
		Transactions: transactionSummary,
	}, nil
}

// importCustomersInBatch imports customers from Excel file in a batch
func (s *ImportServiceImpl) importCustomersInBatch(ctx context.Context, tx *bun.Tx, customerBatchID int, file multipart.File, filename string, batchDate time.Time) (*model.ImportCustomerSummary, *model.ImportLog, error) {
	log := logger.FromContext(ctx, 2)

	rows, err := parser.ReadCustomerExcel(file)
	if err != nil {
		return nil, nil, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to read customer Excel file")
	}

	log.Info().Int("total_rows", len(rows)).Msg("Starting customer import")

	customers := make([]*customerModel.Customer, 0, len(rows))
	errors := []model.ImportRowError{}

	for i, row := range rows {
		normalized := parser.NormalizeName(row.NamaPelanggan)
		phone := row.NomorTelepon

		if normalized == "" {
			errors = append(errors, model.ImportRowError{RowNumber: i + 1, Field: "NamaPelanggan", Message: "Customer name is empty"})
			continue
		}

		customers = append(customers, &customerModel.Customer{
			Name:  normalized,
			Phone: phone,
		})
	}

	newCount, updatedCount, err := s.customerService.BulkImportCustomers(ctx, customers)
	if err != nil {
		return nil, nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to bulk import customers")
	}

	importLog := &model.ImportLog{
		ImportType:      constant.ImportTypeCustomer.String(),
		FileDate:        batchDate,
		Filename:        filename,
		RowsImported:    newCount + updatedCount,
		Status:          constant.ImportStatusSuccess.String(),
		CustomerBatchID: &customerBatchID,
	}
	createdLog, err := s.importLogRepo.Create(ctx, tx, importLog)
	if err != nil {
		return nil, nil, err
	}

	summary := &model.ImportCustomerSummary{
		TotalRows:        len(rows),
		SuccessRows:      newCount + updatedCount,
		FailedRows:       len(errors),
		CustomersCreated: newCount,
		CustomersUpdated: updatedCount,
		Errors:           errors,
	}

	return summary, createdLog, nil
}

// importTransactionsInBatch imports transactions from Excel file with batched commits
func (s *ImportServiceImpl) importTransactionsInBatch(ctx context.Context, transactionBatchID int, file multipart.File, filename string, batchDate time.Time) (*model.ImportTransactionSummary, *model.ImportLog, []int, error) {
	log := logger.FromContext(ctx, 2)

	rows, err := parser.ReadTransactionExcel(file)
	if err != nil {
		return nil, nil, []int{}, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to read transaction Excel file")
	}

	log.Info().Int("total_rows", len(rows)).Msg("Starting transaction import")

	totalRows := len(rows)
	successRows := 0
	failedRows := 0
	transactionsCreated := make(map[string]bool)
	detailsCreated := 0
	errors := []model.ImportRowError{}
	uniqueCustomerIDs := make(map[int]bool)

	var tx *bun.Tx
	txValue, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, []int{}, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start transaction")
	}
	tx = &txValue
	defer tx.Rollback()

	for _, row := range rows {
		if (successRows+failedRows) > 0 && (successRows+failedRows)%100 == 0 {
			if err := tx.Commit(); err != nil {
				log.Error().Err(err).Msg("Failed to commit batch")
				return nil, nil, []int{}, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit batch")
			}

			txValue, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				return nil, nil, []int{}, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start new transaction")
			}
			tx = &txValue
		}

		isNewTransaction, customerID, err := s.processTransactionRow(ctx, tx, row)

		if err != nil {
			failedRows++
			errors = append(errors, model.ImportRowError{RowNumber: row.No + 1, Message: err.Error()})
			log.Error().Int("row", row.No+1).Err(err).Msg("Failed to process transaction row")
			continue
		}

		successRows++
		detailsCreated++

		if isNewTransaction {
			transactionsCreated[row.NoStruk] = true
		}

		if customerID != nil && *customerID > 0 {
			uniqueCustomerIDs[*customerID] = true
		}

		if (successRows+failedRows)%100 == 0 {
			log.Info().Int("processed", successRows+failedRows).Int("total", totalRows).Msg("Progress update")
		}
	}

	importLog := &model.ImportLog{
		ImportType:         constant.ImportTypeTransaction.String(),
		FileDate:           batchDate,
		Filename:           filename,
		RowsImported:       successRows,
		Status:             constant.ImportStatusSuccess.String(),
		TransactionBatchID: &transactionBatchID,
	}
	createdLog, err := s.importLogRepo.Create(ctx, tx, importLog)
	if err != nil {
		return nil, nil, []int{}, err
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit final batch")
		return nil, nil, []int{}, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit final batch")
	}

	log.Info().Int("transaction_import_log_id", createdLog.ID).Msg("Transaction import log created")
	log.Info().Int("total", totalRows).Int("success", successRows).Int("failed", failedRows).Int("transactions", len(transactionsCreated)).Int("details", detailsCreated).Msg("Transaction import completed")

	customerIDs := make([]int, 0, len(uniqueCustomerIDs))
	for id := range uniqueCustomerIDs {
		customerIDs = append(customerIDs, id)
	}

	summary := &model.ImportTransactionSummary{
		TotalRows:                 totalRows,
		SuccessRows:               successRows,
		FailedRows:                failedRows,
		TransactionsCreated:       len(transactionsCreated),
		TransactionDetailsCreated: detailsCreated,
		ProductsCreated:           0,
		VariantsCreated:           0,
		Errors:                    errors,
	}
	return summary, createdLog, customerIDs, nil
}

// processTransactionRow processes a single transaction row
func (s *ImportServiceImpl) processTransactionRow(ctx context.Context, tx *bun.Tx, row *model.TransactionExcelRow) (bool, *int, error) {
	// 1. Try to find customer by name
	var customerID *int
	normalizedName := parser.NormalizeName(row.NamaPelanggan)
	customer, err := s.customerService.GetCustomerByName(ctx, normalizedName)
	if err == nil && customer != nil {
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

	product, err = s.productService.GetOrCreateProductInTx(ctx, tx, product)
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

		variant, err = s.variantService.GetOrCreateVariantInTx(ctx, tx, variant)
		if err != nil {
			return false, nil, fmt.Errorf("failed to get/create variant: %w", err)
		}
		variantID = &variant.ID
	}

	// 4. Parse transaction date
	transactionDate, err := parser.ParseTransactionDateWithTimezone(row.Tanggal, row.Jam)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse transaction date: %w", err)
	}

	paymentMethod, err := parser.NormalizePaymentMethod(row.MetodePembayaran)
	if err != nil {
		return false, nil, fmt.Errorf("failed to normalize payment method: %w", err)
	}
	existingTransaction, _ := s.transactionService.GetTransactionByCodeInTx(ctx, tx, row.NoStruk)
	isNewTransaction := (existingTransaction == nil)

	if isNewTransaction {
		var guestName *string
		if customerID == nil {
			guestName = &normalizedName
		}

		transaction := &transactionModel.Transaction{
			Code:            row.NoStruk,
			CustomerID:      customerID,
			GuestName:       guestName,
			TransactionDate: transactionDate,
			Discount:        row.DiskonTransaksi,
			ShippingCost:    row.OngkosKirim,
			PaymentMethod:   paymentMethod,
			Status:          row.Status,
		}

		if err := s.transactionService.CreateTransactionInTx(ctx, tx, transaction); err != nil {
			return false, nil, fmt.Errorf("failed to create transaction: %w", err)
		}
	}

	// 7. Calculate unit price and subtotal
	unitPrice := product.BasePrice + parsedVariant.PriceModifier
	subtotal := unitPrice * float64(row.JumlahProduk)

	if subtotal != row.Subtotal {
		return false, nil, fmt.Errorf("subtotal mismatch: calculated %.2f, got %.2f", subtotal, row.Subtotal)
	}

	// 8. Create transaction detail
	detail := &transactionModel.TransactionDetail{
		TransactionCode: row.NoStruk,
		ProductID:       product.ID,
		VariantID:       variantID,
		Quantity:        row.JumlahProduk,
		UnitPrice:       unitPrice,
		Subtotal:        subtotal,
	}

	if err := s.transactionDetailService.CreateTransactionDetailInTx(ctx, tx, detail); err != nil {
		return false, nil, fmt.Errorf("failed to create transaction detail: %w", err)
	}

	return isNewTransaction, customerID, nil
}
