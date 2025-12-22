package service

import (
	"context"
	"fmt"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	customerModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/constant"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/parser"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products"
	productsModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions"
	transactionsModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

// ========== UPDATE: STRUCT - TAMBAH 2 FIELD BARU ==========
type ImportServiceImpl struct {
	db                       *bun.DB
	customerService          customer.CustomerService
	productService           products.ProductService
	variantService           products.VariantService
	transactionService       transactions.TransactionService
	transactionDetailService transactions.TransactionDetailService
	customerBatchRepo        repository.CustomerBatchRepository
	transactionBatchRepo     repository.TransactionBatchRepository
	importLogRepo            repository.ImportLogRepository
	importTrackerRepo        repository.ImportTrackerRepository     // ← BARU
	predictionOrchestrator   customer.PredictionOrchestratorService // ← BARU
}

// ========== UPDATE: CONSTRUCTOR - TAMBAH 2 PARAM BARU ==========
func NewImportService(
	db *bun.DB,
	customerService customer.CustomerService,
	productService products.ProductService,
	variantService products.VariantService,
	transactionService transactions.TransactionService,
	transactionDetailService transactions.TransactionDetailService,
	customerBatchRepo repository.CustomerBatchRepository,
	transactionBatchRepo repository.TransactionBatchRepository,
	importLogRepo repository.ImportLogRepository,
	importTrackerRepo repository.ImportTrackerRepository, // ← BARU
	predictionOrchestrator customer.PredictionOrchestratorService, // ← BARU
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
		importTrackerRepo:        importTrackerRepo,      // ← BARU
		predictionOrchestrator:   predictionOrchestrator, // ← BARU
	}
}

// ========== UPDATE: ImportBatch METHOD ==========
func (s *ImportServiceImpl) ImportBatch(ctx context.Context, req *model.ImportBatchRequest) (*model.ImportBatchResponse, error) {
	log := logger.FromContext(ctx, 2)

	log.Info().Msg("Starting batch import process")

	// ========== BARU: PARSE DATE RANGE DARI REQUEST ==========
	importStartDate, importEndDate, err := req.GetParsedDates()
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse date range from request")
		return nil, response.WrapAppError(ctx, err, response.ErrBadRequest, "Invalid date format")
	}

	log.Info().
		Str("start_date", importStartDate.Format("2006-01-02")).
		Str("end_date", importEndDate.Format("2006-01-02")).
		Msg("Import date range parsed")

	// ========== BARU: VALIDATE SEQUENTIAL IMPORT ==========
	tracker, err := s.importTrackerRepo.GetLatest(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get import tracker")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to validate import sequence")
	}

	if tracker != nil && tracker.ID > 0 {
		expectedStartDate := tracker.LastImportEndDate.AddDate(0, 0, 1)

		if !importStartDate.Equal(expectedStartDate) {
			errMsg := fmt.Sprintf("Import gap detected. Expected start_date: %s, got: %s. Missing data from %s to %s",
				expectedStartDate.Format("2006-01-02"),
				importStartDate.Format("2006-01-02"),
				expectedStartDate.Format("2006-01-02"),
				importStartDate.AddDate(0, 0, -1).Format("2006-01-02"),
			)
			log.Error().Msg(errMsg)
			return nil, response.WrapAppError(ctx, fmt.Errorf(errMsg), response.ErrBadRequest, "Sequential import validation failed")
		}

		log.Info().
			Str("last_import_end", tracker.LastImportEndDate.Format("2006-01-02")).
			Str("current_start", importStartDate.Format("2006-01-02")).
			Msg("Sequential validation passed")
	} else {
		log.Info().Msg("First import detected, skipping sequential validation")
	}
	// ========== END BARU ==========

	// ========== EXISTING: Extract batch dates from filenames ==========
	var customerBatchDate, transactionBatchDate time.Time

	if req.FileCustomer != nil {
		customerBatchDate, err = parser.ExtractBatchDateFromFilename(req.FileCustomer.Filename, constant.ImportTypeCustomer)
		if err != nil {
			log.Error().Err(err).Str("filename", req.FileCustomer.Filename).Msg("Failed to parse customer batch date")
			return nil, response.WrapAppError(ctx, err, response.ErrBadRequest, "Invalid customer filename format")
		}
	}

	transactionBatchDate, err = parser.ExtractBatchDateFromFilename(req.FileTransaction.Filename, constant.ImportTypeTransaction)
	if err != nil {
		log.Error().Err(err).Str("filename", req.FileTransaction.Filename).Msg("Failed to parse transaction batch date")
		return nil, response.WrapAppError(ctx, err, response.ErrBadRequest, "Invalid transaction filename format")
	}

	// Validate: transaction date <= customer date (if customer file exists)
	if req.FileCustomer != nil && transactionBatchDate.After(customerBatchDate) {
		log.Error().Msg("Transaction batch date cannot be after customer batch date")
		return nil, response.WrapAppError(ctx, fmt.Errorf("transaction date (%s) > customer date (%s)",
			transactionBatchDate.Format("2006-01-02"),
			customerBatchDate.Format("2006-01-02"),
		), response.ErrBadRequest, "Invalid batch dates")
	}

	// ========== EXISTING: TX1 - Import customer batch ==========
	var customerBatch *model.CustomerBatch

	if req.FileCustomer != nil {
		log.Info().Msg("Starting customer import transaction")

		err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Create customer batch (inactive initially)
			customerBatch = &model.CustomerBatch{
				BatchDate:  customerBatchDate,
				Filename:   req.FileCustomer.Filename,
				ImportedAt: time.Now(),
				IsActive:   false,
				Notes:      req.Notes,
			}

			_, err := s.customerBatchRepo.Create(ctx, tx, customerBatch)
			if err != nil {
				return err
			}

			log.Info().Int("batch_id", customerBatch.ID).Msg("Customer batch created")

			// Open and read customer file
			file, err := req.FileCustomer.Open()
			if err != nil {
				log.Error().Err(err).Msg("Failed to open customer file")
				return response.WrapAppError(ctx, err, response.ErrInternalError, "Failed to open customer file")
			}
			defer file.Close()

			customerRows, err := parser.ReadCustomerExcel(file)
			if err != nil {
				log.Error().Err(err).Msg("Failed to read customer Excel")
				return response.WrapAppError(ctx, err, response.ErrBadRequest, "Failed to parse customer Excel")
			}

			log.Info().Int("row_count", len(customerRows)).Msg("Customer rows parsed")

			// Parse and normalize customers
			var customers []*customerModel.Customer
			for _, row := range customerRows {
				parsedInfo := parser.NormalizeCustomer(row.NamaPelanggan, row.NomorTelepon)

				customers = append(customers, &customerModel.Customer{
					Name:  parsedInfo.Name,
					Phone: parsedInfo.Phone,
				})
			}

			// Bulk import customers
			result, err := s.customerService.BulkImportCustomers(ctx, customers)
			if err != nil {
				log.Error().Err(err).Msg("Failed to bulk import customers")
				return err
			}

			// Update customer batch counts
			customerBatch.CustomerCount = result.TotalProcessed
			customerBatch.NewCustomers = result.NewCustomers
			customerBatch.UpdatedCustomers = result.UpdatedCustomers
			customerBatch.UpgradedFromGuest = result.UpgradedFromGuest

			_, err = s.customerBatchRepo.Update(ctx, tx, customerBatch)
			if err != nil {
				return err
			}

			// Create import log
			importLog := &model.ImportLog{
				ImportType:      constant.ImportTypeCustomer,
				FileDate:        customerBatchDate,
				Filename:        req.FileCustomer.Filename,
				RowsImported:    result.TotalProcessed,
				Status:          constant.ImportStatusSuccess,
				ImportedAt:      time.Now(),
				CustomerBatchID: &customerBatch.ID,
			}

			_, err = s.importLogRepo.Create(ctx, tx, importLog)
			if err != nil {
				return err
			}

			log.Info().
				Int("new", result.NewCustomers).
				Int("updated", result.UpdatedCustomers).
				Int("upgraded", result.UpgradedFromGuest).
				Msg("Customer import completed")

			return nil
		})

		if err != nil {
			log.Error().Err(err).Msg("Customer import transaction failed")
			return nil, err
		}
	} else {
		log.Info().Msg("No customer file provided, skipping customer import")
	}

	// ========== EXISTING: Create transaction batch ==========
	transactionBatch := &model.TransactionBatch{
		BatchDate:  transactionBatchDate,
		Filename:   req.FileTransaction.Filename,
		ImportedAt: time.Now(),
		Notes:      req.Notes,
	}

	if customerBatch != nil {
		transactionBatch.CustomerBatchID = &customerBatch.ID
	}

	_, err = s.transactionBatchRepo.Create(ctx, s.db, transactionBatch)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create transaction batch")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create transaction batch")
	}

	log.Info().Int("batch_id", transactionBatch.ID).Msg("Transaction batch created")

	// ========== EXISTING: Open and read transaction file ==========
	file, err := req.FileTransaction.Open()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open transaction file")
		return nil, response.WrapAppError(ctx, err, response.ErrInternalError, "Failed to open transaction file")
	}
	defer file.Close()

	transactionRows, err := parser.ReadTransactionExcel(file)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read transaction Excel")
		return nil, response.WrapAppError(ctx, err, response.ErrBadRequest, "Failed to parse transaction Excel")
	}

	log.Info().Int("row_count", len(transactionRows)).Msg("Transaction rows parsed")

	// ========== BARU: VALIDATE TRANSACTION DATES DALAM RANGE ==========
	for i, row := range transactionRows {
		txDate, err := parser.ParseTransactionDateWithTimezone(row.Tanggal, row.Jam)
		if err != nil {
			log.Error().Err(err).Int("row", i+1).Msg("Invalid transaction date format")
			return nil, response.WrapAppError(ctx, err, response.ErrBadRequest, fmt.Sprintf("Row %d: invalid date format", i+1))
		}

		txDateOnly := time.Date(txDate.Year(), txDate.Month(), txDate.Day(), 0, 0, 0, 0, time.UTC)

		if txDateOnly.Before(importStartDate) || txDateOnly.After(importEndDate) {
			errMsg := fmt.Sprintf("Row %d: transaction date %s is outside import range [%s, %s]",
				i+1,
				txDateOnly.Format("2006-01-02"),
				importStartDate.Format("2006-01-02"),
				importEndDate.Format("2006-01-02"))
			log.Error().Msg(errMsg)
			return nil, response.WrapAppError(ctx, fmt.Errorf(errMsg), response.ErrBadRequest, "Transaction date out of range")
		}
	}
	log.Info().Msg("All transaction dates validated within range")
	// ========== END BARU ==========

	// ========== EXISTING: Process transactions in batches ==========
	batchSize := 100
	totalRows := len(transactionRows)
	registeredCustomerIDs := make(map[int]bool)

	log.Info().Int("batch_size", batchSize).Msg("Starting transaction processing")

	for i := 0; i < totalRows; i += batchSize {
		end := i + batchSize
		if end > totalRows {
			end = totalRows
		}

		batch := transactionRows[i:end]
		batchNum := (i / batchSize) + 1

		log.Debug().Int("batch", batchNum).Int("rows", len(batch)).Msg("Processing batch")

		err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			for rowIdx, row := range batch {
				globalRowNum := i + rowIdx + 1

				isNew, customerID, err := s.processTransactionRow(ctx, tx, row, transactionBatch.ID, globalRowNum)
				if err != nil {
					log.Error().Err(err).Int("row", globalRowNum).Msg("Failed to process transaction row")
					return err
				}

				if isNew {
					transactionBatch.TransactionCount++
					if customerID > 0 {
						transactionBatch.RegisteredTransactions++
						registeredCustomerIDs[customerID] = true
					} else {
						transactionBatch.GuestTransactions++
					}
				}
			}

			return nil
		})

		if err != nil {
			log.Error().Err(err).Int("batch", batchNum).Msg("Batch processing failed")
			return nil, err
		}

		log.Info().Int("batch", batchNum).Int("processed", end).Int("total", totalRows).Msg("Batch completed")
	}

	log.Info().
		Int("total_transactions", transactionBatch.TransactionCount).
		Int("registered", transactionBatch.RegisteredTransactions).
		Int("guest", transactionBatch.GuestTransactions).
		Int("unique_customers", len(registeredCustomerIDs)).
		Msg("All transactions processed")

	// ========== EXISTING: Update transaction batch ==========
	_, err = s.transactionBatchRepo.Update(ctx, s.db, transactionBatch)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update transaction batch")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to update transaction batch")
	}

	// Create import log for transactions
	importLog := &model.ImportLog{
		ImportType:         constant.ImportTypeTransaction,
		FileDate:           transactionBatchDate,
		Filename:           req.FileTransaction.Filename,
		RowsImported:       transactionBatch.TransactionCount,
		Status:             constant.ImportStatusSuccess,
		ImportedAt:         time.Now(),
		TransactionBatchID: &transactionBatch.ID,
	}

	if customerBatch != nil {
		importLog.CustomerBatchID = &customerBatch.ID
	}

	_, err = s.importLogRepo.Create(ctx, s.db, importLog)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create transaction import log")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create import log")
	}

	// ========== EXISTING: Set customer batch as active ==========
	if customerBatch != nil {
		err = s.customerBatchRepo.SetActive(ctx, s.db, customerBatch.ID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to set customer batch active")
			return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to set customer batch active")
		}
		log.Info().Int("batch_id", customerBatch.ID).Msg("Customer batch set as active")
	}

	// ========== BARU: PROCESS PREDICTIONS ==========
	log.Info().Msg("Processing customer predictions...")

	err = s.predictionOrchestrator.ProcessPredictions(ctx, importStartDate, importEndDate, transactionBatch.ID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to process predictions (non-critical)")
		// Don't fail the entire import
	} else {
		log.Info().Msg("Predictions processed successfully")
	}
	// ========== END BARU ==========

	// ========== EXISTING: Build response ==========
	response := &model.ImportBatchResponse{
		Batch: model.BatchInfo{
			CustomerBatchID:    nil,
			TransactionBatchID: transactionBatch.ID,
			ImportedAt:         time.Now(),
		},
		Transactions: model.ImportTransactionSummary{
			TotalTransactions:      transactionBatch.TransactionCount,
			RegisteredTransactions: transactionBatch.RegisteredTransactions,
			GuestTransactions:      transactionBatch.GuestTransactions,
			UniqueCustomers:        len(registeredCustomerIDs),
		},
	}

	if customerBatch != nil {
		response.Batch.CustomerBatchID = &customerBatch.ID
		response.Customers = &model.ImportCustomerSummary{
			TotalProcessed:    customerBatch.CustomerCount,
			NewCustomers:      customerBatch.NewCustomers,
			UpdatedCustomers:  customerBatch.UpdatedCustomers,
			UpgradedFromGuest: customerBatch.UpgradedFromGuest,
		}
	}

	log.Info().Msg("Batch import completed successfully")

	return response, nil
}

// ========== EXISTING: processTransactionRow (NO CHANGE) ==========
func (s *ImportServiceImpl) processTransactionRow(
	ctx context.Context,
	tx bun.Tx,
	row *model.TransactionExcelRow,
	transactionBatchID int,
	rowNumber int,
) (isNewTransaction bool, customerID int, err error) {
	log := logger.FromContext(ctx, 2)

	// 1. Find customer by name
	normalizedName := parser.NormalizeCustomer(row.NamaPelanggan, "").Name
	customer, err := s.customerService.FindOrCreateCustomerWithNameMatching(ctx, tx, normalizedName, "")
	if err != nil {
		log.Error().Err(err).Str("customer_name", normalizedName).Msg("Failed to find customer")
		return false, 0, err
	}

	var customerIDPtr *int
	var guestName string

	if customer != nil {
		customerIDPtr = &customer.ID
		customerID = customer.ID
	} else {
		guestName = normalizedName
	}

	// 2. Parse product
	parsedProduct, err := parser.ParseProduct(row.NamaProduk)
	if err != nil {
		log.Error().Err(err).Str("product_name", row.NamaProduk).Msg("Failed to parse product")
		return false, 0, fmt.Errorf("row %d: %w", rowNumber, err)
	}

	// 3. Get or create product
	product, err := s.productService.GetOrCreateProductInTx(ctx, tx, &productsModel.Product{
		Name:      parsedProduct.ProductInfo.Name,
		Category:  string(parsedProduct.ProductInfo.Category),
		BasePrice: parsedProduct.ProductInfo.BasePrice,
	})
	if err != nil {
		log.Error().Err(err).Str("product_name", parsedProduct.ProductInfo.Name).Msg("Failed to get/create product")
		return false, 0, fmt.Errorf("row %d: %w", rowNumber, err)
	}

	// 4. Handle variant if product has variants
	var variantIDPtr *int
	var priceModifier float64

	if parsedProduct.ProductInfo.HasVariants {
		parsedVariant, err := parser.ParseVariant(
			row.NamaProduk,
			row.Varian,
			row.HargaVarian,
			row.JumlahProduk,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse variant")
			return false, 0, fmt.Errorf("row %d: %w", rowNumber, err)
		}

		variant, err := s.variantService.GetOrCreateVariantInTx(ctx, tx, &productsModel.Variant{
			ProductID:     product.ID,
			Name:          string(parsedVariant.Size),
			PriceModifier: parsedVariant.PriceModifier,
			IsDefault:     parsedVariant.IsDefault,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get/create variant")
			return false, 0, fmt.Errorf("row %d: %w", rowNumber, err)
		}

		variantIDPtr = &variant.ID
		priceModifier = parsedVariant.PriceModifier
	}

	// 5. Parse transaction date
	transactionDate, err := parser.ParseTransactionDateWithTimezone(row.Tanggal, row.Jam)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse transaction date")
		return false, 0, fmt.Errorf("row %d: %w", rowNumber, err)
	}

	// 6. Check if transaction already exists
	existingTx, err := s.transactionService.GetTransactionByCodeInTx(ctx, tx, row.NoStruk)
	if err != nil {
		return false, 0, fmt.Errorf("row %d: failed to check existing transaction: %w", rowNumber, err)
	}

	var transaction *transactionsModel.Transaction

	if existingTx == nil {
		// Create new transaction
		transaction = &transactionsModel.Transaction{
			Code:            row.NoStruk,
			CustomerID:      customerIDPtr,
			GuestName:       guestName,
			TransactionDate: transactionDate,
			Discount:        row.DiskonTransaksi,
			ShippingCost:    row.OngkosKirim,
			PaymentMethod:   row.MetodePembayaran,
			Status:          row.Status,
		}

		transaction, err = s.transactionService.CreateTransactionInTx(ctx, tx, transaction)
		if err != nil {
			log.Error().Err(err).Str("code", row.NoStruk).Msg("Failed to create transaction")
			return false, 0, fmt.Errorf("row %d: %w", rowNumber, err)
		}

		isNewTransaction = true
	} else {
		transaction = existingTx
		isNewTransaction = false
	}

	// 7. Calculate unit price and validate subtotal
	unitPrice := product.BasePrice + priceModifier
	calculatedSubtotal := unitPrice * float64(row.JumlahProduk)

	if calculatedSubtotal != row.Subtotal {
		log.Warn().
			Float64("calculated", calculatedSubtotal).
			Float64("actual", row.Subtotal).
			Msg("Subtotal mismatch")
	}

	// 8. Create transaction detail
	detail := &transactionsModel.TransactionDetail{
		TransactionCode:    transaction.Code,
		ProductID:          product.ID,
		VariantID:          variantIDPtr,
		Quantity:           row.JumlahProduk,
		UnitPrice:          unitPrice,
		Subtotal:           row.Subtotal,
		TransactionBatchID: transactionBatchID,
	}

	_, err = s.transactionDetailService.CreateTransactionDetailInTx(ctx, tx, detail)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create transaction detail")
		return false, 0, fmt.Errorf("row %d: %w", rowNumber, err)
	}

	return isNewTransaction, customerID, nil
}
