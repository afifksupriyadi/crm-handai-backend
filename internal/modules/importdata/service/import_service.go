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
	productSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products"
	productModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	transactionSvc "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions"
	transactionModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
)

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
	}
}

func (s *ImportServiceImpl) ImportBatch(ctx context.Context, customerFile, transactionFile multipart.File, customerFilename, transactionFilename, notes string) (*model.ImportBatchResponse, error) {
	log := logger.FromContext(ctx, 2)

	transactionDate, err := parser.ExtractBatchDateFromFilename(transactionFilename, string(constant.ImportTypeTransaction))
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidFilenameFormat, err.Error())
	}

	var customerDate time.Time
	if customerFile != nil {
		customerDate, err = parser.ExtractBatchDateFromFilename(customerFilename, string(constant.ImportTypeCustomer))
		if err != nil {
			return nil, response.WrapAppError(ctx, err, response.ErrInvalidFilenameFormat, err.Error())
		}

		if transactionDate.After(customerDate) {
			return nil, response.WrapAppError(ctx, nil, response.ErrTransactionDateExceedsCustomer, fmt.Sprintf("Transaction date (%s) cannot exceed customer date (%s)", transactionDate.Format("2006-01-02"), customerDate.Format("2006-01-02")))
		}
	} else {
		customerDate = transactionDate
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start transaction")
	}
	defer tx.Rollback()

	log.Info().Str("customer_date", customerDate.Format("2006-01-02")).Str("transaction_date", transactionDate.Format("2006-01-02")).Bool("has_customer_file", customerFile != nil).Msg("Starting batch import")

	customerBatch := &model.CustomerBatch{BatchDate: customerDate, Filename: customerFilename, Notes: notes, IsActive: false}
	createdCustomerBatch, err := s.customerBatchRepo.Create(ctx, &tx, customerBatch)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to create customer batch")
	}

	log.Info().Int("customer_batch_id", createdCustomerBatch.ID).Msg("Customer batch created")

	var customerSummary *model.ImportCustomerSummary
	var customerImportLog *model.ImportLog

	if customerFile != nil {
		log.Info().Msg("Importing customers from file")
		customerSummary, customerImportLog, err = s.importCustomersInBatch(ctx, &tx, createdCustomerBatch.ID, customerFile, customerFilename, customerDate)
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
		customerSummary = &model.ImportCustomerSummary{TotalRows: 0, SuccessRows: 0, FailedRows: 0, CustomersCreated: 0, CustomersUpdated: 0, Errors: []model.ImportRowError{}}
	}

	transactionBatch := &model.TransactionBatch{BatchDate: transactionDate, Filename: transactionFilename, CustomerBatchID: createdCustomerBatch.ID, Notes: notes}
	createdTransactionBatch, err := s.transactionBatchRepo.Create(ctx, &tx, transactionBatch)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to create transaction batch")
	}

	log.Info().Int("transaction_batch_id", createdTransactionBatch.ID).Msg("Transaction batch created")

	// Commit main transaction before batched import
	if err := tx.Commit(); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit main transaction")
	}

	transactionSummary, transactionImportLog, customerIDs, err := s.importTransactionsInBatch(ctx, createdTransactionBatch.ID, transactionFile, transactionFilename, transactionDate)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to import transactions")
	}
	log.Info().Int("transaction_import_log_id", transactionImportLog.ID).Msg("Transaction import log created")

	// Start new transaction for final updates
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

	log.Info().Int("customer_batch_id", createdCustomerBatch.ID).Int("transaction_batch_id", createdTransactionBatch.ID).Msg("Batch import committed, starting analytics")

	if err := s.computeCustomerAnalytics(ctx, customerIDs, createdTransactionBatch.ID); err != nil {
		log.Warn().Err(err).Msg("Failed to compute batch analytics")
	} else {
		log.Info().Msg("Batch analytics computed successfully")
	}
	log.Info().Msg("Batch import completed successfully")

	return &model.ImportBatchResponse{
		Batch: &model.BatchInfo{
			ID:        createdTransactionBatch.ID,
			BatchCode: fmt.Sprintf("TB_%s", transactionDate.Format("20060102")),
			BatchDate: transactionDate.Format("2006-01-02"),
			Status:    "COMPLETED",
			IsActive:  true,
		},
		Customers:    customerSummary,
		Transactions: transactionSummary,
	}, nil
}

func (s *ImportServiceImpl) importCustomersInBatch(ctx context.Context, tx *bun.Tx, customerBatchID int, file multipart.File, filename string, batchDate time.Time) (*model.ImportCustomerSummary, *model.ImportLog, error) {
	log := logger.FromContext(ctx, 2)

	rows, err := parser.ReadCustomerExcel(file)
	if err != nil {
		return nil, nil, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to read customer Excel file")
	}

	log.Info().Int("total_rows", len(rows)).Msg("Starting customer import")

	customers := make([]*customerModel.Customer, 0, len(rows))
	errors := []model.ImportRowError{}

	for _, row := range rows {
		parsed, err := parser.NormalizeCustomer(row.NamaPelanggan, row.NomorTelepon)
		if err != nil {
			errors = append(errors, model.ImportRowError{RowNumber: row.No + 1, Message: err.Error()})
			log.Error().Int("row", row.No+1).Err(err).Msg("Failed to normalize customer")
			continue
		}

		customers = append(customers, &customerModel.Customer{Name: parsed.Name, Phone: parsed.Phone})
	}

	newCount, updatedCount, err := s.customerService.BulkImportCustomers(ctx, customers)
	if err != nil {
		return nil, nil, response.WrapAppError(ctx, err, response.ErrBatchProcessing, "Failed to bulk import customers")
	}

	successRows := newCount + updatedCount
	failedRows := len(rows) - successRows

	importLog := &model.ImportLog{
		ImportType:      constant.ImportTypeCustomer.String(),
		FileDate:        batchDate,
		Filename:        filename,
		RowsImported:    successRows,
		Status:          constant.ImportStatusSuccess.String(),
		CustomerBatchID: &customerBatchID,
	}
	createdLog, err := s.importLogRepo.Create(ctx, tx, importLog)
	if err != nil {
		return nil, nil, err
	}

	log.Info().Int("total", len(rows)).Int("success", successRows).Int("failed", failedRows).Int("created", newCount).Int("updated", updatedCount).Msg("Customer import completed")

	summary := &model.ImportCustomerSummary{TotalRows: len(rows), SuccessRows: successRows, FailedRows: failedRows, CustomersCreated: newCount, CustomersUpdated: updatedCount, Errors: errors}
	return summary, createdLog, nil
}

func (s *ImportServiceImpl) importTransactionsInBatch(ctx context.Context, transactionBatchID int, file multipart.File, filename string, batchDate time.Time) (*model.ImportTransactionSummary, *model.ImportLog, []int, error) {
	log := logger.FromContext(ctx, 2)

	rows, err := parser.ReadTransactionExcel(file)
	if err != nil {
		return nil, nil, []int{}, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to read transaction Excel file")
	}

	log.Info().Int("total_rows", len(rows)).Msg("Starting transaction import")

	var totalRows = len(rows)
	var successRows = 0
	var failedRows = 0
	var transactionsCreated = make(map[string]bool)
	var detailsCreated = 0
	var errors []model.ImportRowError
	var batchSize = 100

	// Track unique customer IDs for analytics
	uniqueCustomerIDs := make(map[int]bool)

	txStruct, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, []int{}, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start first transaction")
	}
	tx := &txStruct
	defer tx.Rollback()

	for i, row := range rows {
		if i%batchSize == 0 && i > 0 {
			if err := tx.Commit(); err != nil {
				log.Error().Err(err).Int("row", i).Msg("Failed to commit batch")
				return nil, nil, []int{}, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit batch")
			}

			log.Info().Int("processed", i).Int("total", totalRows).Msg("Batch committed, starting new transaction")

			newTxStruct, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				log.Error().Err(err).Int("row", i).Msg("Failed to start new transaction")
				return nil, nil, []int{}, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start new transaction")
			}
			tx = &newTxStruct
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

	// 5. Check if transaction exists, create if new
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
			PaymentMethod:   row.MetodePembayaran,
			Status:          row.Status,
		}

		if customerID == nil {
			transaction.GuestName = &normalizedName
		}

		if err := s.transactionService.CreateTransactionInTx(ctx, tx, transaction); err != nil {
			return false, nil, fmt.Errorf("failed to create transaction: %w", err)
		}
	}

	// 6. Calculate unit price and subtotal
	unitPrice := product.BasePrice + parsedVariant.PriceModifier
	subtotal := unitPrice * float64(row.JumlahProduk)

	if subtotal != row.Subtotal {
		return false, nil, fmt.Errorf("subtotal mismatch: calculated %.2f, got %.2f", subtotal, row.Subtotal)
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
		return false, nil, fmt.Errorf("failed to create transaction detail: %w", err)
	}

	return isNewTransaction, customerID, nil
}

func (s *ImportServiceImpl) computeCustomerAnalytics(ctx context.Context, customerIDs []int, transactionBatchID int) error {
	log := logger.FromContext(ctx, 2)

	if len(customerIDs) == 0 {
		log.Info().Msg("No registered customers in this batch, skipping analytics")
		return nil
	}

	log.Info().Int("total_customers", len(customerIDs)).Int("transaction_batch_id", transactionBatchID).Msg("Computing analytics for registered customers")

	successCount := 0
	for _, customerID := range customerIDs {
		if err := s.customerService.ComputeCustomerMetrics(ctx, customerID, transactionBatchID); err != nil {
			log.Warn().Err(err).Int("customer_id", customerID).Msg("Failed to compute metrics")
			continue
		}
		successCount++
	}

	log.Info().Int("success", successCount).Int("total", len(customerIDs)).Msg("Analytics computation completed")
	return nil
}
