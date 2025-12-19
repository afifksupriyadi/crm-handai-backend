package repository

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

type TransactionDetailRepository interface {
	GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error)
	GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error)
	CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error
	CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error
	CreateTransactionDetailInTx(ctx context.Context, tx *bun.Tx, detail *model.TransactionDetail) error
}

type TransactionDetailRepositoryImpl struct {
	db *bun.DB
}

func NewTransactionDetailRepository(db *bun.DB) TransactionDetailRepository {
	return &TransactionDetailRepositoryImpl{db: db}
}

// Regular methods
func (r *TransactionDetailRepositoryImpl) GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error) {
	return r.getTransactionDetailByID(ctx, r.db, id)
}

func (r *TransactionDetailRepositoryImpl) GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error) {
	return r.getTransactionDetailsByTransactionCode(ctx, r.db, transactionCode)
}

func (r *TransactionDetailRepositoryImpl) CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error {
	return r.createTransactionDetail(ctx, r.db, detail)
}

func (r *TransactionDetailRepositoryImpl) CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error {
	return r.createTransactionDetailsBulk(ctx, r.db, details)
}

// Transaction methods
func (r *TransactionDetailRepositoryImpl) CreateTransactionDetailInTx(ctx context.Context, tx *bun.Tx, detail *model.TransactionDetail) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}
	return r.createTransactionDetail(ctx, db, detail)
}

// Internal shared implementations
func (r *TransactionDetailRepositoryImpl) getTransactionDetailByID(ctx context.Context, db bun.IDB, id int) (*model.TransactionDetail, error) {
	detail := new(model.TransactionDetail)
	err := db.NewSelect().Model(detail).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find transaction detail by ID")
		return nil, err
	}
	return detail, nil
}

func (r *TransactionDetailRepositoryImpl) getTransactionDetailsByTransactionCode(ctx context.Context, db bun.IDB, transactionCode string) ([]*model.TransactionDetail, error) {
	var details []*model.TransactionDetail
	err := db.NewSelect().Model(&details).Where("transaction_code = ?", transactionCode).Where("deleted_at IS NULL").Scan(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find transaction details by code")
		return nil, err
	}
	return details, nil
}

func (r *TransactionDetailRepositoryImpl) createTransactionDetail(ctx context.Context, db bun.IDB, detail *model.TransactionDetail) error {
	_, err := db.NewInsert().Model(detail).Exec(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create transaction detail")
		return err
	}
	return nil
}

func (r *TransactionDetailRepositoryImpl) createTransactionDetailsBulk(ctx context.Context, db bun.IDB, details []*model.TransactionDetail) error {
	_, err := db.NewInsert().Model(&details).Exec(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create transaction details bulk")
		return err
	}
	return nil
}
