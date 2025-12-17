package repository

import (
	"context"

	"github.com/uptrace/bun"
)

type ImportRepository interface {
	GetCustomerIDsInBatch(ctx context.Context, batchID int) ([]int, error)
}

type ImportRepositoryImpl struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) ImportLogRepository {
	return &ImportLogRepositoryImpl{db: db}
}

// GetCustomerIDsInBatch returns all customer IDs that have transactions in a batch
func (r *ImportRepositoryImpl) GetCustomerIDsInBatch(ctx context.Context, batchID int) ([]int, error) {
	var customerIDs []int

	err := r.db.NewSelect().
		Table("transactions").
		ColumnExpr("DISTINCT customer_id").
		Where("batch_id = ?", batchID).
		Where("deleted_at IS NULL").
		Scan(ctx, &customerIDs)

	return customerIDs, err
}
