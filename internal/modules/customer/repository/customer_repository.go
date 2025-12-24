package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	transactionModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

type CustomerRepositoryImpl struct {
	db *bun.DB
}

// NewCustomerRepository creates a new instance of CustomerRepositoryImpl
func NewCustomerRepository(db *bun.DB) customer.CustomerRepository {
	return &CustomerRepositoryImpl{db: db}
}

// Create inserts a new customer into the database
func (r *CustomerRepositoryImpl) Create(ctx context.Context, db bun.IDB, customer *model.Customer) (*model.Customer, error) {
	customer.CreatedAt = time.Now()

	_, err := db.NewInsert().
		Model(customer).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create customer")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create customer")
	}

	return customer, nil
}

// FindByID retrieves a customer by ID
func (r *CustomerRepositoryImpl) FindByID(ctx context.Context, db bun.IDB, id int) (*model.Customer, error) {
	customer := new(model.Customer)

	err := db.NewSelect().
		Model(customer).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, response.WrapAppError(ctx, err, response.ErrCustomerNotFound, "Customer not found")
		}
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find customer by ID")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to find customer")
	}

	return customer, nil
}

// FindByPhone retrieves a customer by phone number
func (r *CustomerRepositoryImpl) FindByPhone(ctx context.Context, db bun.IDB, phone string) (*model.Customer, error) {
	customer := new(model.Customer)

	err := db.NewSelect().
		Model(customer).
		Where("phone = ?", phone).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find customer by phone")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to find customer")
	}

	return customer, nil
}

// FindByName retrieves a customer by name (case-insensitive)
func (r *CustomerRepositoryImpl) FindByName(ctx context.Context, db bun.IDB, name string) (*model.Customer, error) {
	customer := new(model.Customer)

	err := db.NewSelect().
		Model(customer).
		Where("LOWER(name) = LOWER(?)", name).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find customer by name")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to find customer")
	}

	return customer, nil
}

// FindAll retrieves all customers with pagination and search
func (r *CustomerRepositoryImpl) FindAll(ctx context.Context, page, limit int, search, sortOrder string) ([]*model.CustomerWithLatestMetrics, int, error) {
	var customersWithMetrics []*model.CustomerWithLatestMetrics

	// Subquery: Latest metric per customer (using ROW_NUMBER)
	latestMetrics := r.db.NewSelect().
		ColumnExpr("customer_id").
		ColumnExpr("last_transaction_date").
		ColumnExpr("is_loyal").
		ColumnExpr("ROW_NUMBER() OVER (PARTITION BY customer_id ORDER BY computed_at DESC) as rn").
		TableExpr("analytics.customer_metrics")

	// Subquery: Latest prediction per customer
	latestPredictions := r.db.NewSelect().
		ColumnExpr("customer_id").
		ColumnExpr("predicted_next_purchase_date").
		ColumnExpr("ROW_NUMBER() OVER (PARTITION BY customer_id ORDER BY created_at DESC) as rn").
		TableExpr("analytics.customer_predictions")

	// Main query
	query := r.db.NewSelect().
		ColumnExpr("c.*").
		ColumnExpr("lm.last_transaction_date").
		ColumnExpr("COALESCE(lm.is_loyal, false) as is_loyal").
		ColumnExpr("cs.segment").
		ColumnExpr("lp.predicted_next_purchase_date").
		TableExpr("customers c").
		// LEFT JOIN with latest metrics (only rn = 1)
		Join("LEFT JOIN (?) lm ON lm.customer_id = c.id AND lm.rn = 1", latestMetrics).
		// LEFT JOIN with segments
		Join("LEFT JOIN analytics.customer_segments cs ON cs.customer_id = c.id").
		// LEFT JOIN with latest predictions (only rn = 1)
		Join("LEFT JOIN (?) lp ON lp.customer_id = c.id AND lp.rn = 1", latestPredictions).
		Where("c.deleted_at IS NULL")

	// Apply search filter
	if search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", search)
		query = query.Where("c.name ILIKE ? OR c.phone ILIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	totalCount, err := r.db.NewSelect().
		TableExpr("customers c").
		Where("c.deleted_at IS NULL").
		Count(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to count customers")
		return nil, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to count customers")
	}

	// Apply sorting
	orderBy := "c.id ASC"
	if strings.ToLower(sortOrder) == "desc" {
		orderBy = "c.id DESC"
	}

	// Apply pagination
	offset := (page - 1) * limit
	err = query.
		Order(orderBy).
		Limit(limit).
		Offset(offset).
		Scan(ctx, &customersWithMetrics)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get customers")
		return nil, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get customers")
	}

	return customersWithMetrics, totalCount, nil
}

// Update updates an existing customer
func (r *CustomerRepositoryImpl) Update(ctx context.Context, db bun.IDB, customer *model.Customer) (*model.Customer, error) {
	now := time.Now()
	customer.UpdatedAt = &now

	result, err := db.NewUpdate().
		Model(customer).
		Column("name", "phone", "updated_at").
		Where("id = ?", customer.ID).
		Where("deleted_at IS NULL").
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to update customer")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to update customer")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, response.WrapAppError(ctx, nil, response.ErrCustomerNotFound, "Customer not found")
	}

	return customer, nil
}

// Delete soft deletes a customer by setting deleted_at timestamp
func (r *CustomerRepositoryImpl) Delete(ctx context.Context, db bun.IDB, id int) error {
	now := time.Now()

	result, err := db.NewUpdate().
		Model((*model.Customer)(nil)).
		Set("deleted_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to delete customer")
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to delete customer")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return response.WrapAppError(ctx, nil, response.ErrCustomerNotFound, "Customer not found")
	}

	return nil
}

// Exists checks if a customer exists by ID
func (r *CustomerRepositoryImpl) Exists(ctx context.Context, db bun.IDB, id int) (bool, error) {
	count, err := db.NewSelect().
		Model((*model.Customer)(nil)).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Count(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to check customer existence")
		return false, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check customer existence")
	}

	return count > 0, nil
}

// WithTx executes function within a transaction
func (r *CustomerRepositoryImpl) WithTx(ctx context.Context, fn func(*bun.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(&tx); err != nil {
		return err
	}

	return tx.Commit()
}

// CountPastGuestTransactions counts past guest transactions without linking
func (r *CustomerRepositoryImpl) CountPastGuestTransactions(ctx context.Context, db bun.IDB, guestName string) (int, error) {
	count, err := db.NewSelect().
		Model((*transactionModel.Transaction)(nil)).
		Where("guest_name = ?", guestName).
		Where("customer_id IS NULL").
		Where("deleted_at IS NULL").
		Count(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Str("guest_name", guestName).Msg("Failed to count past guest transactions")
		return 0, err
	}

	return count, nil
}

// LinkPastTransactions links past guest transactions to newly registered customer
func (r *CustomerRepositoryImpl) LinkPastTransactions(ctx context.Context, db bun.IDB, guestName string, customerID int) (int, error) {
	result, err := db.NewUpdate().
		Model((*transactionModel.Transaction)(nil)).
		Set("customer_id = ?", customerID).
		SetColumn("guest_name", "NULL").
		Set("updated_at = NOW()").
		Where("guest_name = ?", guestName).
		Where("customer_id IS NULL").
		Where("deleted_at IS NULL").
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Str("guest_name", guestName).Msg("Failed to link past transactions")
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

func (r *CustomerRepositoryImpl) ComputeAndStoreMetrics(ctx context.Context, db bun.IDB, customerID int, transactionBatchID int) error {
	var stats struct {
		TotalTransactions   int
		TotalSpent          float64
		LastTransactionDate time.Time
	}

	err := db.NewSelect().
		TableExpr("transactions t").
		ColumnExpr("COUNT(*) as total_transactions").
		ColumnExpr(`SUM((SELECT COALESCE(SUM(subtotal), 0) FROM transaction_details WHERE transaction_code = t.code) - t.discount + t.shipping_cost) as total_spent`).
		ColumnExpr("MAX(t.transaction_date) as last_transaction_date").
		Where("t.customer_id = ?", customerID).
		Where("t.deleted_at IS NULL").
		Scan(ctx, &stats)

	if err != nil {
		return fmt.Errorf("failed to query transaction stats: %w", err)
	}

	var avgDays *float64
	if stats.TotalTransactions > 1 {
		var dates []time.Time
		err := db.NewSelect().
			Table("transactions").
			Column("transaction_date").
			Where("customer_id = ?", customerID).
			Where("deleted_at IS NULL").
			Order("transaction_date ASC").
			Scan(ctx, &dates)

		if err != nil {
			return fmt.Errorf("failed to query transaction dates: %w", err)
		}

		if len(dates) > 1 {
			totalDays := dates[len(dates)-1].Sub(dates[0]).Hours() / 24
			avg := totalDays / float64(len(dates)-1)
			avgDays = &avg
		}
	}

	// Get is_loyal status from customer_segments table
	var segment string
	var isLoyal bool
	err = db.NewSelect().
		TableExpr("analytics.customer_segments").
		Column("segment").
		Where("customer_id = ?", customerID).
		Scan(ctx, &segment)

	if err == nil && segment == "LOYAL" {
		isLoyal = true
	}

	metric := &model.CustomerMetric{
		CustomerID:             customerID,
		TransactionBatchID:     transactionBatchID,
		TotalTransactions:      stats.TotalTransactions,
		TotalSpent:             stats.TotalSpent,
		LastTransactionDate:    &stats.LastTransactionDate,
		AvgDaysBetweenPurchase: avgDays,
		IsLoyal:                isLoyal,
	}

	_, err = db.NewInsert().
		Model(metric).
		On("CONFLICT (customer_id, transaction_batch_id) DO UPDATE").
		Set("total_transactions = EXCLUDED.total_transactions").
		Set("total_spent = EXCLUDED.total_spent").
		Set("last_transaction_date = EXCLUDED.last_transaction_date").
		Set("avg_days_between_purchase = EXCLUDED.avg_days_between_purchase").
		Set("is_loyal = EXCLUDED.is_loyal").
		Set("computed_at = EXCLUDED.computed_at").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to insert customer metrics: %w", err)
	}

	return nil
}

// FindAllWithRecentTransactions retrieves customers with their latest metrics from analytics schema
func (r *CustomerRepositoryImpl) FindAllWithRecentTransactions(ctx context.Context, page, limit int) ([]*model.CustomerWithMetrics, int, error) {
	var results []*model.CustomerWithMetrics

	// Subquery to get latest metrics per customer
	latestMetricsSubquery := r.db.NewSelect().
		TableExpr("analytics.customer_metrics").
		Column("customer_id").
		ColumnExpr("MAX(computed_at) as max_computed_at").
		Group("customer_id")

	// Main query: JOIN customers with their latest metrics + segments
	query := r.db.NewSelect().
		ColumnExpr("c.*").                     // Customer columns
		ColumnExpr("cm.transaction_batch_id"). // Metrics columns
		ColumnExpr("cm.total_transactions").
		ColumnExpr("cm.total_spent").
		ColumnExpr("cm.last_transaction_date").
		ColumnExpr("cm.avg_days_between_purchase").
		ColumnExpr("cm.segment"). // ← Keep this (redundant but harmless)
		ColumnExpr("cm.is_loyal").
		ColumnExpr("cm.churn_risk_score").
		ColumnExpr("cm.computed_at").
		ColumnExpr("cs.segment as segment_status"). // ← NEW: Explicit alias from customer_segments
		TableExpr("customers c").
		Join("INNER JOIN analytics.customer_metrics cm ON cm.customer_id = c.id").
		Join("INNER JOIN (?) lm ON lm.customer_id = cm.customer_id AND cm.computed_at = lm.max_computed_at", latestMetricsSubquery).
		Join("LEFT JOIN analytics.customer_segments cs ON cs.customer_id = c.id"). // ← NEW: Join with segments
		Where("c.deleted_at IS NULL").
		Where("cm.last_transaction_date IS NOT NULL") // Only customers with transactions

	// Get total count
	totalCount, err := query.Count(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to count customers with transactions")
		return nil, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to count customers")
	}

	// Apply pagination and sort by most recent transaction
	offset := (page - 1) * limit
	err = query.
		Order("cm.last_transaction_date DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx, &results)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get customers with recent transactions")
		return nil, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get customers")
	}

	return results, totalCount, nil
}

// GetCustomerDetailData retrieves comprehensive customer data for detail page
func (r *CustomerRepositoryImpl) GetCustomerDetailData(ctx context.Context, customerID int, year *int, month *int) (*model.CustomerDetailData, error) {
	data := &model.CustomerDetailData{}

	// 1. Get basic customer info
	err := r.db.NewSelect().
		Model(&data.Customer).
		Where("id = ?", customerID).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, response.WrapAppError(ctx, err, response.ErrCustomerNotFound, "Customer not found")
		}
		return nil, err
	}

	// 2. Get latest metrics WITH segment JOIN
	data.Metrics = &model.CustomerMetric{}
	err = r.db.NewSelect().
		ColumnExpr("cm.*").
		ColumnExpr("cs.segment").
		TableExpr("analytics.customer_metrics cm").
		Join("LEFT JOIN analytics.customer_segments cs ON cs.customer_id = cm.customer_id").
		Where("cm.customer_id = ?", customerID).
		Order("cm.computed_at DESC").
		Limit(1).
		Scan(ctx, data.Metrics)

	if err != nil && err != sql.ErrNoRows {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get customer metrics")
		data.Metrics = nil
	}
	if err == sql.ErrNoRows {
		data.Metrics = nil
	}

	// 3. Get all transactions with details (with optional year/month filter)
	query := r.db.NewSelect().
		ColumnExpr("t.code").
		ColumnExpr("t.transaction_date").
		ColumnExpr("t.discount").
		ColumnExpr("t.shipping_cost").
		ColumnExpr("p.name as product_name").
		ColumnExpr("v.name as variant_name").
		ColumnExpr("td.quantity").
		ColumnExpr("td.subtotal").
		TableExpr("transactions t").
		Join("INNER JOIN transaction_details td ON td.transaction_code = t.code").
		Join("INNER JOIN products p ON p.id = td.product_id").
		Join("LEFT JOIN variants v ON v.id = td.variant_id").
		Where("t.customer_id = ?", customerID).
		Where("t.deleted_at IS NULL").
		Where("td.deleted_at IS NULL").
		Order("t.transaction_date DESC")

	// Apply filters based on year and month parameters
	if year != nil && month != nil {
		// Filter specific year and month
		query = query.
			Where("EXTRACT(YEAR FROM t.transaction_date) = ?", *year).
			Where("EXTRACT(MONTH FROM t.transaction_date) = ?", *month)
	} else if year != nil {
		// Filter specific year, all months
		query = query.Where("EXTRACT(YEAR FROM t.transaction_date) = ?", *year)
	} else if month != nil {
		// Filter specific month, all years
		query = query.Where("EXTRACT(MONTH FROM t.transaction_date) = ?", *month)
	}
	// If both nil, no filter (all data)

	err = query.Scan(ctx, &data.TransactionDetails)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get transaction details")
		return nil, err
	}

	// 4. Get product aggregates (respect same filters)
	var productAggregates []model.ProductAggregate

	aggregateSQL := `
SELECT 
    p.name as product_name,
    v.name as variant_name,
    SUM(td.quantity) as total_quantity
FROM transaction_details td
INNER JOIN transactions t ON t.code = td.transaction_code
INNER JOIN products p ON p.id = td.product_id
LEFT JOIN variants v ON v.id = td.variant_id
WHERE t.customer_id = ?
AND t.deleted_at IS NULL
AND td.deleted_at IS NULL
`

	args := []interface{}{customerID}

	// Apply filters
	if year != nil && month != nil {
		aggregateSQL += " AND EXTRACT(YEAR FROM t.transaction_date) = ? AND EXTRACT(MONTH FROM t.transaction_date) = ?"
		args = append(args, *year, *month)
	} else if year != nil {
		aggregateSQL += " AND EXTRACT(YEAR FROM t.transaction_date) = ?"
		args = append(args, *year)
	} else if month != nil {
		aggregateSQL += " AND EXTRACT(MONTH FROM t.transaction_date) = ?"
		args = append(args, *month)
	}

	aggregateSQL += `
GROUP BY p.name, v.name
ORDER BY SUM(td.quantity) DESC
`

	err = r.db.NewRaw(aggregateSQL, args...).Scan(ctx, &productAggregates)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get product aggregates")
	} else {
		data.ProductAggregates = productAggregates
	}

	// 5. Get prediction (always latest, no filter)
	data.Prediction = &model.CustomerPrediction{}
	err = r.db.NewSelect().
		Model(data.Prediction).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	if err != nil && err != sql.ErrNoRows {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get customer prediction")
		data.Prediction = nil
	}
	if err == sql.ErrNoRows {
		data.Prediction = nil
	}

	// 6. Get predicted products (if prediction exists)
	if data.Prediction != nil {
		var predictedProducts []*model.CustomerPredictedProduct
		err = r.db.NewSelect().
			Model(&predictedProducts).
			Relation("Product").
			Relation("Variant").
			Where("prediction_id = ?", data.Prediction.ID).
			Order("confidence_score DESC").
			Scan(ctx)

		if err != nil && err != sql.ErrNoRows {
			logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get predicted products")
		}

		data.PredictedProducts = predictedProducts
	}

	return data, nil
}

// GetCustomersBySegment retrieves customers by segment type with metrics
func (r *CustomerRepositoryImpl) GetCustomersBySegment(ctx context.Context, segmentType string, limit int) ([]*model.CustomerWithLatestMetrics, error) {
	var customersWithMetrics []*model.CustomerWithLatestMetrics

	// Subquery: Latest metric per customer
	latestMetrics := r.db.NewSelect().
		ColumnExpr("customer_id").
		ColumnExpr("last_transaction_date").
		ColumnExpr("is_loyal").
		ColumnExpr("ROW_NUMBER() OVER (PARTITION BY customer_id ORDER BY computed_at DESC) as rn").
		TableExpr("analytics.customer_metrics")

	// Subquery: Transaction count per customer
	transactionCounts := r.db.NewSelect().
		ColumnExpr("customer_id").
		ColumnExpr("COUNT(*) as transaction_count").
		TableExpr("transactions").
		Where("deleted_at IS NULL").
		Group("customer_id")

	// Main query: Join customers with segments and latest metrics
	err := r.db.NewSelect().
		ColumnExpr("c.*").
		ColumnExpr("lm.last_transaction_date").
		ColumnExpr("COALESCE(lm.is_loyal, false) as is_loyal").
		ColumnExpr("cs.segment").
		TableExpr("customers c").
		Join("INNER JOIN analytics.customer_segments cs ON cs.customer_id = c.id").
		Join("LEFT JOIN (?) lm ON lm.customer_id = c.id AND lm.rn = 1", latestMetrics).
		Join("LEFT JOIN (?) tc ON tc.customer_id = c.id", transactionCounts).
		Where("c.deleted_at IS NULL").
		Where("cs.segment = ?", segmentType).
		Order("tc.transaction_count DESC NULLS LAST").
		Limit(limit).
		Scan(ctx, &customersWithMetrics)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Str("segment", segmentType).Msg("Failed to get customers by segment")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get customers by segment")
	}

	return customersWithMetrics, nil
}

// GetCustomerPurchaseCounts returns weekly and monthly product purchase counts
func (r *CustomerRepositoryImpl) GetCustomerPurchaseCounts(ctx context.Context, customerID int) (weeklyCount int, monthlyCount int) {
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)

	// Query result struct
	type CountResult struct {
		Total int `bun:"total"`
	}

	// Weekly purchases (sum of quantities in last 7 days)
	var weeklyResult CountResult
	err := r.db.NewSelect().
		TableExpr("transaction_details td").
		ColumnExpr("COALESCE(SUM(td.quantity), 0) as total").
		Join("INNER JOIN transactions t ON t.code = td.transaction_code").
		Where("t.customer_id = ?", customerID).
		Where("t.transaction_date >= ?", weekAgo).
		Where("t.deleted_at IS NULL").
		Scan(ctx, &weeklyResult)

	if err == nil {
		weeklyCount = weeklyResult.Total
	}

	// Monthly purchases (sum of quantities in last 30 days)
	var monthlyResult CountResult
	err = r.db.NewSelect().
		TableExpr("transaction_details td").
		ColumnExpr("COALESCE(SUM(td.quantity), 0) as total").
		Join("INNER JOIN transactions t ON t.code = td.transaction_code").
		Where("t.customer_id = ?", customerID).
		Where("t.transaction_date >= ?", monthAgo).
		Where("t.deleted_at IS NULL").
		Scan(ctx, &monthlyResult)

	if err == nil {
		monthlyCount = monthlyResult.Total
	}

	return weeklyCount, monthlyCount
}

// GetCustomerTransactionCount returns total number of transactions for a customer
func (r *CustomerRepositoryImpl) GetCustomerTransactionCount(ctx context.Context, customerID int) int {
	type CountResult struct {
		Total int `bun:"total"`
	}

	var result CountResult
	err := r.db.NewSelect().
		TableExpr("transactions").
		ColumnExpr("COUNT(*) as total").
		Where("customer_id = ?", customerID).
		Where("deleted_at IS NULL").
		Scan(ctx, &result)

	if err != nil {
		return 0
	}

	return result.Total
}

func (r *CustomerRepositoryImpl) GetCustomerPredictions(ctx context.Context, customerID int, limit int) ([]*model.CustomerPrediction, error) {
	var predictions []*model.CustomerPrediction

	err := r.db.NewSelect().
		Model(&predictions).
		Relation("PredictedProducts").         // ← NEW
		Relation("PredictedProducts.Product"). // ← NEW
		Relation("PredictedProducts.Variant"). // ← NEW
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Msg("Failed to get customer predictions")
		return nil, err
	}

	return predictions, nil
}
