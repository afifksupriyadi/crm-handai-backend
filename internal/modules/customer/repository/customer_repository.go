package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/constant"
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
func (r *CustomerRepositoryImpl) FindAll(ctx context.Context, page, limit int, search, sortOrder string) ([]*model.Customer, int, error) {
	var customers []*model.Customer

	query := r.db.NewSelect().
		Model(&customers).
		Where("deleted_at IS NULL")

	if search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", search)
		query = query.Where("name ILIKE ? OR phone ILIKE ?", searchPattern, searchPattern)
	}

	totalCount, err := query.Count(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to count customers")
		return nil, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to count customers")
	}

	orderBy := "id ASC"
	if strings.ToLower(sortOrder) == "desc" {
		orderBy = "id DESC"
	}

	offset := (page - 1) * limit
	err = query.
		Order(orderBy).
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get customers")
		return nil, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get customers")
	}

	return customers, totalCount, nil
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

// ComputeAndStoreMetrics calculates customer metrics and stores to analytics.customer_metrics
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

	segment := determineCustomerSegment(stats.TotalTransactions, avgDays, stats.LastTransactionDate)
	churnRisk := calculateChurnRisk(stats.LastTransactionDate, avgDays)

	metric := &model.CustomerMetric{
		CustomerID:             customerID,
		TransactionBatchID:     transactionBatchID,
		TotalTransactions:      stats.TotalTransactions,
		TotalSpent:             stats.TotalSpent,
		LastTransactionDate:    &stats.LastTransactionDate,
		AvgDaysBetweenPurchase: avgDays,
		Segment:                &segment,
		IsLoyal:                segment == string(constant.SegmentLoyal),
		ChurnRiskScore:         churnRisk,
	}

	_, err = db.NewInsert().
		Model(metric).
		On("CONFLICT (customer_id, transaction_batch_id) DO UPDATE").
		Set("total_transactions = EXCLUDED.total_transactions").
		Set("total_spent = EXCLUDED.total_spent").
		Set("last_transaction_date = EXCLUDED.last_transaction_date").
		Set("avg_days_between_purchase = EXCLUDED.avg_days_between_purchase").
		Set("segment = EXCLUDED.segment").
		Set("is_loyal = EXCLUDED.is_loyal").
		Set("churn_risk_score = EXCLUDED.churn_risk_score").
		Set("computed_at = EXCLUDED.computed_at").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to insert customer metrics: %w", err)
	}

	return nil
}

func determineCustomerSegment(totalTx int, avgDays *float64, lastTxDate time.Time) string {
	daysSinceLast := time.Since(lastTxDate).Hours() / 24

	if totalTx <= 2 {
		return string(constant.SegmentNew)
	}
	if totalTx <= 5 {
		return string(constant.SegmentPotential)
	}
	if avgDays != nil && daysSinceLast > (*avgDays*2) {
		return string(constant.SegmentChurn)
	}
	return string(constant.SegmentLoyal)
}

func calculateChurnRisk(lastTxDate time.Time, avgDays *float64) *float64 {
	if avgDays == nil || *avgDays == 0 {
		return nil
	}

	daysSinceLast := time.Since(lastTxDate).Hours() / 24
	risk := daysSinceLast / *avgDays

	if risk > 1.0 {
		risk = 1.0
	}

	return &risk
}
