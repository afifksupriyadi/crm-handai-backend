package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

type CustomerServiceImpl struct {
	repo customer.CustomerRepository
	db   *bun.DB
}

// NewCustomerService creates a new instance of CustomerServiceImpl
func NewCustomerService(repo customer.CustomerRepository, db *bun.DB) customer.CustomerService {
	return &CustomerServiceImpl{
		repo: repo,
		db:   db,
	}
}

// CreateCustomer creates a new customer
func (s *CustomerServiceImpl) CreateCustomer(ctx context.Context, req *model.CreateCustomerRequest) (*model.CustomerResponse, error) {
	log := logger.FromContext(ctx, 2)

	existingCustomer, err := s.repo.FindByPhone(ctx, s.db, req.Phone)
	if err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to check existing customer")
		return nil, err
	}

	if existingCustomer != nil {
		return nil, response.WrapAppError(ctx, nil, response.ErrPhoneAlreadyExists, "Phone number already exists")
	}

	customer := &model.Customer{Name: req.Name, Phone: req.Phone}
	createdCustomer, err := s.repo.Create(ctx, s.db, customer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create customer")
		return nil, err
	}

	log.Info().Int("customer_id", createdCustomer.ID).Msg("Customer created successfully")
	return mapCustomerToResponse(createdCustomer), nil
}

// GetCustomerByID retrieves a customer by ID
func (s *CustomerServiceImpl) GetCustomerByID(ctx context.Context, id int) (*model.CustomerResponse, error) {
	log := logger.FromContext(ctx, 2)

	customer, err := s.repo.FindByID(ctx, s.db, id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to get customer by ID")
		return nil, err
	}

	return mapCustomerToResponse(customer), nil
}

// GetCustomerByPhone retrieves a customer by phone number
func (s *CustomerServiceImpl) GetCustomerByPhone(ctx context.Context, phone string) (*model.Customer, error) {
	customer, err := s.repo.FindByPhone(ctx, s.db, phone)
	if err != nil {
		logger.FromContext(ctx, 2).Error().Err(err).Str("phone", phone).Msg("Failed to get customer by phone")
		return nil, err
	}
	return customer, nil
}

// GetCustomerByName retrieves a customer by name
func (s *CustomerServiceImpl) GetCustomerByName(ctx context.Context, name string) (*model.Customer, error) {
	customer, err := s.repo.FindByName(ctx, s.db, name)
	if err != nil {
		logger.FromContext(ctx, 2).Error().Err(err).Str("name", name).Msg("Failed to get customer by name")
		return nil, err
	}
	return customer, nil
}

// GetOrCreateCustomer gets or creates a customer by name and phone
func (s *CustomerServiceImpl) GetOrCreateCustomer(ctx context.Context, name, phone string) (*model.Customer, error) {
	log := logger.FromContext(ctx, 2)

	customer, err := s.repo.FindByPhone(ctx, s.db, phone)
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("Failed to find customer by phone")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to find customer")
	}

	if customer != nil {
		log.Info().Int("customer_id", customer.ID).Msg("Customer found by phone")
		return customer, nil
	}

	newCustomer := &model.Customer{Name: name, Phone: phone}
	createdCustomer, err := s.repo.Create(ctx, s.db, newCustomer)
	if err != nil {
		log.Error().Err(err).Str("name", name).Str("phone", phone).Msg("Failed to create customer")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create customer")
	}

	log.Info().Int("customer_id", createdCustomer.ID).Msg("Customer created via import")
	return createdCustomer, nil
}

func (s *CustomerServiceImpl) GetAllCustomers(ctx context.Context, req *model.GetCustomersRequest) (*model.CustomerListResponse, error) {
	log := logger.FromContext(ctx, 2)

	customers, totalCount, err := s.repo.FindAll(ctx, req.Page, req.Limit, req.Search, req.SortOrder)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all customers")
		return nil, err
	}

	now := time.Now()
	customerResponses := make([]*model.CustomerResponse, len(customers))
	for i, customer := range customers {
		resp := &model.CustomerResponse{
			ID:                  customer.ID,
			Name:                customer.Name,
			Phone:               customer.Phone,
			Status:              customer.Segment, // NEW
			LastTransactionDate: customer.LastTransactionDate,
			SupposeToByBy:       customer.PredictedNextPurchaseDate, // NEW
			CreatedAt:           customer.CreatedAt,
			UpdatedAt:           customer.UpdatedAt,
			UpgradedFromGuest:   customer.UpgradedFromGuest,
			UpgradedAt:          customer.UpgradedAt,
			FirstSeenAsGuest:    customer.FirstSeenAsGuest,
			IsLoyal:             customer.IsLoyal,
		}

		// Calculate days since last purchase if applicable
		if customer.LastTransactionDate != nil {
			daysSince := int(now.Sub(*customer.LastTransactionDate).Hours() / 24)
			resp.DaysSinceLastPurchase = &daysSince
		}

		customerResponses[i] = resp
	}

	totalPages := (totalCount + req.Limit - 1) / req.Limit

	return &model.CustomerListResponse{
		Data: customerResponses,
		Pagination: model.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// UpdateCustomer updates an existing customer
func (s *CustomerServiceImpl) UpdateCustomer(ctx context.Context, id int, req *model.UpdateCustomerRequest) (*model.CustomerResponse, error) {
	log := logger.FromContext(ctx, 2)

	customer, err := s.repo.FindByID(ctx, s.db, id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to find customer for update")
		return nil, err
	}

	if req.Phone != customer.Phone {
		existingWithPhone, err := s.repo.FindByPhone(ctx, s.db, req.Phone)
		if err != nil {
			log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to check phone existence")
			return nil, err
		}
		if existingWithPhone != nil {
			return nil, response.WrapAppError(ctx, nil, response.ErrPhoneAlreadyExists, "Phone number already exists")
		}
	}

	customer.Name = req.Name
	customer.Phone = req.Phone

	updatedCustomer, err := s.repo.Update(ctx, s.db, customer)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to update customer")
		return nil, err
	}

	log.Info().Int("customer_id", updatedCustomer.ID).Msg("Customer updated successfully")
	return mapCustomerToResponse(updatedCustomer), nil
}

// DeleteCustomer soft deletes a customer
func (s *CustomerServiceImpl) DeleteCustomer(ctx context.Context, id int) error {
	log := logger.FromContext(ctx, 2)

	_, err := s.repo.FindByID(ctx, s.db, id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Customer not found for deletion")
		return err
	}

	err = s.repo.Delete(ctx, s.db, id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to delete customer")
		return err
	}

	log.Info().Int("customer_id", id).Msg("Customer deleted successfully")
	return nil
}

// FindOrCreateCustomerWithNameMatching finds or creates customer with fuzzy name matching
func (s *CustomerServiceImpl) FindOrCreateCustomerWithNameMatching(ctx context.Context, tx *bun.Tx, name, phone string) (*model.Customer, bool, error) {
	log := logger.FromContext(ctx, 2)

	existing, err := s.repo.FindByPhone(ctx, tx, phone)
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("Failed to get customer by phone")
		return nil, false, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check existing customer")
	}

	if existing != nil {
		if namesAreSimilar(existing.Name, name) {
			log.Debug().Int("customer_id", existing.ID).Str("existing_name", existing.Name).Str("import_name", name).Msg("Customer matched by phone and similar name")
			return existing, false, nil
		}

		log.Warn().Int("customer_id", existing.ID).Str("old_name", existing.Name).Str("new_name", name).Msg("Updating customer name due to mismatch")
		existing.Name = name
		existing.UpdatedAt = timePtr(time.Now())
		_, err = s.repo.Update(ctx, tx, existing)
		if err != nil {
			log.Error().Err(err).Int("id", existing.ID).Msg("Failed to update customer name")
			return nil, false, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to update customer")
		}
		return existing, false, nil
	}

	newCustomer := &model.Customer{Name: name, Phone: phone}
	createdCustomer, err := s.repo.Create(ctx, tx, newCustomer)
	if err != nil {
		log.Error().Err(err).Str("name", name).Str("phone", phone).Msg("Failed to create customer")
		return nil, false, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create customer")
	}

	log.Info().Int("customer_id", createdCustomer.ID).Str("name", name).Msg("Customer created in batch import")
	return createdCustomer, true, nil
}

// BulkImportCustomers imports multiple customers from Excel and handles retroactive linking
func (s *CustomerServiceImpl) BulkImportCustomers(ctx context.Context, customers []*model.Customer) (int, int, error) {
	log := logger.FromContext(ctx, 2)

	var newCount, updatedCount int

	err := s.repo.WithTx(ctx, func(tx *bun.Tx) error {
		for _, customer := range customers {
			existing, err := s.repo.FindByPhone(ctx, tx, customer.Phone)
			if err != nil {
				log.Error().Err(err).Str("phone", customer.Phone).Msg("Failed to check existing customer")
				continue
			}

			if existing != nil {
				existing.Name = customer.Name
				existing.UpdatedAt = timePtr(time.Now())
				_, err = s.repo.Update(ctx, tx, existing)
				if err != nil {
					log.Error().Err(err).Int("id", existing.ID).Msg("Failed to update customer")
					continue
				}
				updatedCount++
				log.Info().Int("customer_id", existing.ID).Msg("Customer updated")
			} else {
				linkedCount, err := s.repo.LinkPastTransactions(ctx, tx, customer.Name, 0)
				if err != nil {
					log.Error().Err(err).Str("name", customer.Name).Msg("Failed to check past transactions")
				}

				if linkedCount > 0 {
					customer.UpgradedFromGuest = true
					customer.UpgradedAt = timePtr(time.Now())
					customer.FirstSeenAsGuest = timePtr(time.Now())
					log.Info().Str("name", customer.Name).Int("linked_count", linkedCount).Msg("Found past guest transactions")
				}

				createdCustomer, err := s.repo.Create(ctx, tx, customer)
				if err != nil {
					log.Error().Err(err).Str("name", customer.Name).Msg("Failed to insert customer")
					continue
				}

				if linkedCount > 0 {
					finalLinked, err := s.repo.LinkPastTransactions(ctx, tx, customer.Name, createdCustomer.ID)
					if err != nil {
						log.Error().Err(err).Int("customer_id", createdCustomer.ID).Msg("Failed to link past transactions")
					} else {
						log.Info().Int("customer_id", createdCustomer.ID).Int("linked", finalLinked).Msg("Guest upgraded and transactions linked")
					}
				}

				newCount++
				log.Info().Int("customer_id", createdCustomer.ID).Msg("Customer created")
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to bulk import customers")
		return 0, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to bulk import customers")
	}

	log.Info().Int("new", newCount).Int("updated", updatedCount).Msg("Bulk import completed")
	return newCount, updatedCount, nil
}

// UpgradeGuestToCustomer upgrades guest to registered customer and links past transactions
func (s *CustomerServiceImpl) UpgradeGuestToCustomer(ctx context.Context, guestName string, customer *model.Customer) (int, error) {
	log := logger.FromContext(ctx, 2)

	var linkedCount int
	err := s.repo.WithTx(ctx, func(tx *bun.Tx) error {
		customer.UpgradedFromGuest = true
		customer.UpgradedAt = timePtr(time.Now())
		customer.FirstSeenAsGuest = timePtr(time.Now())

		createdCustomer, err := s.repo.Create(ctx, tx, customer)
		if err != nil {
			log.Error().Err(err).Str("name", customer.Name).Msg("Failed to insert customer")
			return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create customer")
		}

		count, err := s.repo.LinkPastTransactions(ctx, tx, guestName, createdCustomer.ID)
		if err != nil {
			log.Error().Err(err).Int("customer_id", createdCustomer.ID).Msg("Failed to link past transactions")
			return err
		}

		linkedCount = count
		log.Info().Int("customer_id", createdCustomer.ID).Int("linked", linkedCount).Msg("Guest upgraded to customer")
		return nil
	})

	return linkedCount, err
}

// mapCustomerToResponse maps Customer model to CustomerResponse DTO
func mapCustomerToResponse(customer *model.Customer) *model.CustomerResponse {
	return &model.CustomerResponse{
		ID:                customer.ID,
		Name:              customer.Name,
		Phone:             customer.Phone,
		CreatedAt:         customer.CreatedAt,
		UpdatedAt:         customer.UpdatedAt,
		UpgradedFromGuest: customer.UpgradedFromGuest,
		UpgradedAt:        customer.UpgradedAt,
		FirstSeenAsGuest:  customer.FirstSeenAsGuest,
	}
}

// namesAreSimilar checks if two names are similar using contains logic
func namesAreSimilar(name1, name2 string) bool {
	n1 := strings.ToLower(strings.TrimSpace(name1))
	n2 := strings.ToLower(strings.TrimSpace(name2))
	return strings.Contains(n1, n2) || strings.Contains(n2, n1)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// ComputeCustomerMetrics computes and stores analytics for a customer
func (s *CustomerServiceImpl) ComputeCustomerMetrics(ctx context.Context, customerID int, transactionBatchID int) error {
	log := logger.FromContext(ctx, 2)

	err := s.repo.ComputeAndStoreMetrics(ctx, s.db, customerID, transactionBatchID)
	if err != nil {
		log.Error().Err(err).Int("customer_id", customerID).Int("transaction_batch_id", transactionBatchID).Msg("Failed to compute customer metrics")
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to compute customer metrics")
	}

	log.Debug().Int("customer_id", customerID).Int("transaction_batch_id", transactionBatchID).Msg("Customer metrics computed")
	return nil
}

// GetCustomersWithRecentTransactions retrieves customers with their latest transaction info from analytics
func (s *CustomerServiceImpl) GetCustomersWithRecentTransactions(ctx context.Context, req *model.GetRecentTransactionsRequest) (*model.CustomerRecentTransactionListResponse, error) {
	log := logger.FromContext(ctx, 2)
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Get customers with metrics from repository
	customersWithMetrics, totalCount, err := s.repo.FindAllWithRecentTransactions(ctx, req.Page, req.Limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get customers with recent transactions")
		return nil, err
	}

	// Map to response DTOs
	now := time.Now()
	customerResponses := make([]*model.CustomerRecentTransactionResponse, 0, len(customersWithMetrics))

	for _, cwm := range customersWithMetrics {
		lastTransactionDateInLocalTime := cwm.CustomerMetric.LastTransactionDate.In(loc)
		resp := &model.CustomerRecentTransactionResponse{
			ID:                     cwm.Customer.ID,
			Name:                   cwm.Customer.Name,
			Phone:                  cwm.Customer.Phone,
			LastTransactionDate:    &lastTransactionDateInLocalTime,
			TotalTransactions:      cwm.CustomerMetric.TotalTransactions,
			TotalSpent:             cwm.CustomerMetric.TotalSpent,
			Segment:                cwm.CustomerMetric.Segment,
			IsLoyal:                cwm.CustomerMetric.IsLoyal,
			AvgDaysBetweenPurchase: cwm.CustomerMetric.AvgDaysBetweenPurchase,
			ChurnRiskScore:         cwm.CustomerMetric.ChurnRiskScore,
		}

		// Calculate days since last transaction
		if cwm.CustomerMetric.LastTransactionDate != nil {
			daysSince := int(now.Sub(*cwm.CustomerMetric.LastTransactionDate).Hours() / 24)
			resp.DaysSinceLastTransaction = &daysSince
		}

		customerResponses = append(customerResponses, resp)
	}

	// Calculate pagination metadata
	totalPages := (totalCount + req.Limit - 1) / req.Limit

	return &model.CustomerRecentTransactionListResponse{
		Data: customerResponses,
		Pagination: model.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *CustomerServiceImpl) GetCustomerDetail(ctx context.Context, id int, month *time.Time) (*model.CustomerDetailResponse, error) {
	log := logger.FromContext(ctx, 2)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	// If no month specified, use current month
	if month == nil {
		month = &now
	}

	// Get comprehensive data from repository
	data, err := s.repo.GetCustomerDetailData(ctx, id, month)
	if err != nil {
		log.Error().Err(err).Int("customer_id", id).Msg("Failed to get customer detail data")
		return nil, err
	}

	// Build response
	resp := &model.CustomerDetailResponse{}

	// 1. Customer Info
	resp.Customer = model.CustomerInfo{
		ID:                data.Customer.ID,
		Name:              data.Customer.Name,
		Phone:             data.Customer.Phone,
		CreatedAt:         data.Customer.CreatedAt,
		UpgradedFromGuest: data.Customer.UpgradedFromGuest,
		UpgradedAt:        data.Customer.UpgradedAt,
		IsLoyal:           false,
	}
	if data.Metrics != nil {
		resp.Customer.Segment = data.Metrics.Segment
		resp.Customer.IsLoyal = data.Metrics.IsLoyal
	}

	// 2. Calculate Stats
	resp.Stats = s.calculateStats(data)

	// 3. Latest Purchase
	resp.LatestPurchase = s.getLatestPurchase(data, now)

	// 4. Prediction - Get latest prediction from new system
	latestPrediction, err := s.repo.GetCustomerPredictions(ctx, id, 1)
	if err == nil && len(latestPrediction) > 0 {
		pred := latestPrediction[0]

		var actualDate *string
		if pred.ActualNextPurchaseDate != nil {
			dateStr := pred.ActualNextPurchaseDate.Format("2006-01-02")
			actualDate = &dateStr
		}

		var daysUntil *int
		if pred.ActualNextPurchaseDate == nil {
			days := int(pred.PredictedNextPurchaseDate.Sub(now).Hours() / 24)
			daysUntil = &days
		}

		resp.Prediction = &model.PredictionInfo{
			LastTransactionDate:       pred.LastTransactionDate.Format("2006-01-02"),
			PredictedNextPurchaseDate: pred.PredictedNextPurchaseDate.Format("2006-01-02"),
			ActualNextPurchaseDate:    actualDate,
			IsPredictedCorrect:        pred.IsPredictedCorrect,
			DaysUntilPredicted:        daysUntil,
			CreatedAt:                 pred.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			PredictedProducts:         s.mapPredictedProducts(data.PredictedProducts),
		}
	} else {
		resp.Prediction = nil
	}

	// 5. Product Purchases (sorted by quantity)
	resp.ProductPurchases = make([]model.ProductPurchaseInfo, 0, len(data.ProductAggregates))
	for _, pa := range data.ProductAggregates {
		resp.ProductPurchases = append(resp.ProductPurchases, model.ProductPurchaseInfo{
			ProductName:   pa.ProductName,
			TotalQuantity: pa.TotalQuantity,
		})
	}

	// 6. Transaction History (grouped by date)
	resp.TransactionHistory = s.buildTransactionHistory(data, month)

	return resp, nil
}

func (s *CustomerServiceImpl) calculateStats(data *model.CustomerDetailData) model.CustomerStats {
	stats := model.CustomerStats{}

	// Calculate from transaction details (already filtered by month in repository)
	productSet := make(map[string]bool)
	totalSpent := 0.0
	transactionSet := make(map[string]bool)

	for _, td := range data.TransactionDetails {
		// Count unique products this month
		productSet[td.ProductName] = true

		// Track unique transactions
		if !transactionSet[td.Code] {
			transactionSet[td.Code] = true
		}
	}

	// Aggregate per transaction for accurate totals
	txMap := make(map[string]struct {
		Subtotals    float64
		Discount     float64
		ShippingCost float64
	})

	for _, td := range data.TransactionDetails {
		tx := txMap[td.Code]
		tx.Subtotals += td.Subtotal
		tx.Discount = td.Discount
		tx.ShippingCost = td.ShippingCost
		txMap[td.Code] = tx
	}

	// Calculate total spent
	for _, tx := range txMap {
		totalSpent += (tx.Subtotals - tx.Discount + tx.ShippingCost)
	}

	stats.TotalProductsPurchases = len(productSet)
	stats.TotalSpent = totalSpent
	stats.TotalTransaction = formatCurrency(totalSpent)

	// Total transaction count from metrics (all time)
	if data.Metrics != nil {
		stats.TotalTransactionCount = data.Metrics.TotalTransactions
	}

	return stats
}

func (s *CustomerServiceImpl) getLatestPurchase(data *model.CustomerDetailData, now time.Time) *model.LatestPurchaseInfo {
	if len(data.TransactionDetails) == 0 {
		return nil
	}

	// Get latest transaction (already sorted DESC by date in repository)
	latestCode := data.TransactionDetails[0].Code
	latestDate := data.TransactionDetails[0].TransactionDate
	daysAgo := int(now.Sub(latestDate).Hours() / 24)

	// Collect products for latest transaction
	products := make([]model.TransactionProductDTO, 0)
	for _, td := range data.TransactionDetails {
		if td.Code == latestCode {
			variant := ""
			if td.VariantName != nil {
				variant = *td.VariantName
			}
			products = append(products, model.TransactionProductDTO{
				Name:     td.ProductName,
				Variant:  variant,
				Quantity: td.Quantity,
			})
		}
	}

	return &model.LatestPurchaseInfo{
		TransactionCode: latestCode,
		Date:            latestDate,
		DaysAgo:         daysAgo,
		Products:        products,
	}
}

func (s *CustomerServiceImpl) buildTransactionHistory(data *model.CustomerDetailData, month *time.Time) model.TransactionHistoryInfo {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Format month string
	filterMonth := month.In(loc).Format("January 2006")

	// Group transactions by date
	txByDate := make(map[string]*model.TransactionHistoryItemDTO)
	txCodes := []string{} // To maintain order

	for _, td := range data.TransactionDetails {
		dateKey := td.TransactionDate.Format("2006-01-02")

		if _, exists := txByDate[dateKey]; !exists {
			txCodes = append(txCodes, td.Code)
			txByDate[dateKey] = &model.TransactionHistoryItemDTO{
				Date:            td.TransactionDate,
				DateFormatted:   td.TransactionDate.In(loc).Format("Monday, 2 Jan 2006"),
				TransactionCode: td.Code,
				Products:        []model.TransactionProductDTO{},
				TotalAmount:     0,
			}
		}

		// Add product
		variant := ""
		if td.VariantName != nil {
			variant = *td.VariantName
		}

		item := txByDate[dateKey]
		item.Products = append(item.Products, model.TransactionProductDTO{
			Name:     td.ProductName,
			Variant:  variant,
			Quantity: td.Quantity,
		})

		// Calculate total (subtotal - discount + shipping)
		// Note: discount & shipping are per transaction, not per item
		// So we need to aggregate properly
	}

	// Fix total calculation - aggregate per transaction code
	txTotals := make(map[string]float64)
	txDiscounts := make(map[string]float64)
	txShipping := make(map[string]float64)

	for _, td := range data.TransactionDetails {
		txTotals[td.Code] += td.Subtotal
		txDiscounts[td.Code] = td.Discount
		txShipping[td.Code] = td.ShippingCost
	}

	// Apply totals
	for _, code := range txCodes {
		dateKey := ""
		for _, td := range data.TransactionDetails {
			if td.Code == code {
				dateKey = td.TransactionDate.Format("2006-01-02")
				break
			}
		}
		if dateKey != "" && txByDate[dateKey] != nil {
			total := txTotals[code] - txDiscounts[code] + txShipping[code]
			txByDate[dateKey].TotalAmount = total
			txByDate[dateKey].TotalFormatted = formatCurrency(total)
		}
	}

	// Convert map to slice
	transactions := make([]model.TransactionHistoryItemDTO, 0, len(txByDate))
	for _, item := range txByDate {
		transactions = append(transactions, *item)
	}

	// Sort by date descending
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.After(transactions[j].Date)
	})

	return model.TransactionHistoryInfo{
		FilterMonth:  filterMonth,
		Transactions: transactions,
	}
}

func formatCurrency(amount float64) string {
	// Format to Indonesian Rupiah
	if amount >= 1000000 {
		return fmt.Sprintf("Rp. %.1fM", amount/1000000)
	} else if amount >= 1000 {
		return fmt.Sprintf("Rp. %.1fK", amount/1000)
	}
	return fmt.Sprintf("Rp. %.0f", amount)
}

// GetCustomerPredictions retrieves prediction history for a customer
func (s *CustomerServiceImpl) GetCustomerPredictions(ctx context.Context, customerID int, limit int) (*model.CustomerPredictionListResponse, error) {
	log := logger.FromContext(ctx, 2)

	// Default limit
	if limit <= 0 || limit > 10 {
		limit = 4
	}

	// Get predictions from repository (this assumes we'll add this method to repo)
	predictions, err := s.repo.GetCustomerPredictions(ctx, customerID, limit)
	if err != nil {
		log.Error().Err(err).Int("customer_id", customerID).Msg("Failed to get customer predictions")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get predictions")
	}

	// Map to response DTOs
	predictionResponses := make([]*model.CustomerPredictionResponse, 0, len(predictions))
	for _, p := range predictions {
		var actualDate *string
		if p.ActualNextPurchaseDate != nil {
			dateStr := p.ActualNextPurchaseDate.Format("2006-01-02")
			actualDate = &dateStr
		}

		predictionResponses = append(predictionResponses, &model.CustomerPredictionResponse{
			ID:                        p.ID,
			LastTransactionDate:       p.LastTransactionDate.Format("2006-01-02"),
			PredictedNextPurchaseDate: p.PredictedNextPurchaseDate.Format("2006-01-02"),
			ActualNextPurchaseDate:    actualDate,
			IsPredictedCorrect:        p.IsPredictedCorrect,
			CreatedAt:                 p.CreatedAt,
			ValidatedAt:               p.ValidatedAt,
			PredictedProducts:         s.mapPredictedProducts(p.PredictedProducts),
		})
	}

	return &model.CustomerPredictionListResponse{
		Data: predictionResponses,
	}, nil
}

func (s *CustomerServiceImpl) mapPredictedProducts(products []*model.CustomerPredictedProduct) []model.PredictedProductInfo {
	result := make([]model.PredictedProductInfo, 0, len(products))

	for _, p := range products {
		var variantName *string
		if p.Variant != nil {
			variantName = &p.Variant.Name
		}

		productName := ""
		if p.Product != nil {
			productName = p.Product.Name
		}

		result = append(result, model.PredictedProductInfo{
			ProductName:       productName,
			VariantName:       variantName,
			PredictedQuantity: p.PredictedQuantity,
			ConfidenceScore:   p.ConfidenceScore,
			PurchaseFrequency: p.PurchaseFrequency,
		})
	}

	return result
}
