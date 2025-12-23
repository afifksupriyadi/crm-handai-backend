package repository

import (
	"context"
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

// CustomerPredictedProductRepository defines the contract for predicted product data access
type CustomerPredictedProductRepository interface {
	// Create creates multiple predicted products for a prediction
	CreateBatch(ctx context.Context, db bun.IDB, products []*model.CustomerPredictedProduct) error

	// GetByPrediction retrieves predicted products for a prediction
	GetByPrediction(ctx context.Context, db bun.IDB, predictionID int) ([]*model.CustomerPredictedProduct, error)

	// GetByPredictionWithDetails retrieves predicted products with product & variant details
	GetByPredictionWithDetails(ctx context.Context, db bun.IDB, predictionID int) ([]*model.CustomerPredictedProduct, error)

	// DeleteByPrediction deletes all predicted products for a prediction (manual, though CASCADE handles it)
	DeleteByPrediction(ctx context.Context, db bun.IDB, predictionID int) error

	GetCustomerTransactionProducts(ctx context.Context, customerID int, transactionLimit int) ([]model.TransactionProduct, error)
}

// CustomerPredictedProductRepositoryImpl implements CustomerPredictedProductRepository interface
type CustomerPredictedProductRepositoryImpl struct {
	db *bun.DB
}

// NewCustomerPredictedProductRepository creates a new CustomerPredictedProductRepository
func NewCustomerPredictedProductRepository(db *bun.DB) CustomerPredictedProductRepository {
	return &CustomerPredictedProductRepositoryImpl{db: db}
}

// CreateBatch creates multiple predicted products in a single transaction
func (r *CustomerPredictedProductRepositoryImpl) CreateBatch(ctx context.Context, db bun.IDB, products []*model.CustomerPredictedProduct) error {
	if len(products) == 0 {
		return nil // Nothing to insert
	}

	_, err := db.NewInsert().
		Model(&products).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().
			Err(err).
			Int("count", len(products)).
			Msg("Failed to create predicted products batch")
		return fmt.Errorf("failed to create predicted products: %w", err)
	}

	logger.FromContext(ctx, 2).Debug().
		Int("count", len(products)).
		Msg("Predicted products created successfully")

	return nil
}

// GetByPrediction retrieves predicted products for a prediction (without relations)
func (r *CustomerPredictedProductRepositoryImpl) GetByPrediction(ctx context.Context, db bun.IDB, predictionID int) ([]*model.CustomerPredictedProduct, error) {
	var products []*model.CustomerPredictedProduct

	err := db.NewSelect().
		Model(&products).
		Where("prediction_id = ?", predictionID).
		Order("confidence_score DESC").
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().
			Err(err).
			Int("prediction_id", predictionID).
			Msg("Failed to get predicted products")
		return nil, fmt.Errorf("failed to get predicted products: %w", err)
	}

	return products, nil
}

// GetByPredictionWithDetails retrieves predicted products with product & variant details
func (r *CustomerPredictedProductRepositoryImpl) GetByPredictionWithDetails(ctx context.Context, db bun.IDB, predictionID int) ([]*model.CustomerPredictedProduct, error) {
	var products []*model.CustomerPredictedProduct

	err := db.NewSelect().
		Model(&products).
		Relation("Product").
		Relation("Variant").
		Where("cpp.prediction_id = ?", predictionID).
		Order("cpp.confidence_score DESC").
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().
			Err(err).
			Int("prediction_id", predictionID).
			Msg("Failed to get predicted products with details")
		return nil, fmt.Errorf("failed to get predicted products with details: %w", err)
	}

	return products, nil
}

// DeleteByPrediction deletes all predicted products for a prediction
// Note: CASCADE DELETE handles this automatically, but keeping for explicit control if needed
func (r *CustomerPredictedProductRepositoryImpl) DeleteByPrediction(ctx context.Context, db bun.IDB, predictionID int) error {
	_, err := db.NewDelete().
		Model((*model.CustomerPredictedProduct)(nil)).
		Where("prediction_id = ?", predictionID).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().
			Err(err).
			Int("prediction_id", predictionID).
			Msg("Failed to delete predicted products")
		return fmt.Errorf("failed to delete predicted products: %w", err)
	}

	return nil
}

// GetCustomerTransactionProducts retrieves product purchase history for prediction analysis
func (r *CustomerPredictedProductRepositoryImpl) GetCustomerTransactionProducts(ctx context.Context, customerID int, transactionLimit int) ([]model.TransactionProduct, error) {
	var txProducts []model.TransactionProduct

	// Get products from last N transactions
	// Fetch more than needed, will filter to exact N transactions in service
	maxRows := transactionLimit * 20 // Assume max 20 items per transaction

	err := r.db.NewSelect().
		ColumnExpr("t.code as transaction_code").
		ColumnExpr("td.product_id").
		ColumnExpr("p.name as product_name").
		ColumnExpr("td.variant_id").
		ColumnExpr("v.name as variant_name").
		ColumnExpr("td.quantity").
		TableExpr("transactions t").
		Join("INNER JOIN transaction_details td ON td.transaction_code = t.code").
		Join("INNER JOIN products p ON p.id = td.product_id").
		Join("LEFT JOIN variants v ON v.id = td.variant_id").
		Where("t.customer_id = ?", customerID).
		Where("t.deleted_at IS NULL").
		Where("td.deleted_at IS NULL").
		Order("t.transaction_date DESC").
		Limit(maxRows).
		Scan(ctx, &txProducts)

	if err != nil {
		logger.FromContext(ctx, 1).Error().
			Err(err).
			Int("customer_id", customerID).
			Msg("Failed to get customer transaction products")
		return nil, fmt.Errorf("failed to get transaction products: %w", err)
	}

	return txProducts, nil
}
