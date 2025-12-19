package repository

import (
	"context"
	"database/sql"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

type TransactionRepository interface {
	GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error)
	CreateTransaction(ctx context.Context, transaction *model.Transaction) error
	GetTransactionByCodeInTx(ctx context.Context, tx *bun.Tx, code string) (*model.Transaction, error)
	CreateTransactionInTx(ctx context.Context, tx *bun.Tx, transaction *model.Transaction) error
}

type TransactionRepositoryImpl struct {
	db *bun.DB
}

func NewTransactionRepository(db *bun.DB) TransactionRepository {
	return &TransactionRepositoryImpl{db: db}
}

// Regular methods
func (r *TransactionRepositoryImpl) GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error) {
	return r.getTransactionByCode(ctx, r.db, code)
}

func (r *TransactionRepositoryImpl) CreateTransaction(ctx context.Context, transaction *model.Transaction) error {
	return r.createTransaction(ctx, r.db, transaction)
}

// Transaction methods
func (r *TransactionRepositoryImpl) GetTransactionByCodeInTx(ctx context.Context, tx *bun.Tx, code string) (*model.Transaction, error) {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}
	return r.getTransactionByCode(ctx, db, code)
}

func (r *TransactionRepositoryImpl) CreateTransactionInTx(ctx context.Context, tx *bun.Tx, transaction *model.Transaction) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}
	return r.createTransaction(ctx, db, transaction)
}

// Internal shared implementations
func (r *TransactionRepositoryImpl) getTransactionByCode(ctx context.Context, db bun.IDB, code string) (*model.Transaction, error) {
	transaction := new(model.Transaction)
	err := db.NewSelect().Model(transaction).Where("code = ?", code).Where("deleted_at IS NULL").Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find transaction by code")
		return nil, err
	}
	return transaction, nil
}

func (r *TransactionRepositoryImpl) createTransaction(ctx context.Context, db bun.IDB, transaction *model.Transaction) error {
	_, err := db.NewInsert().Model(transaction).Exec(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create transaction")
		return err
	}
	return nil
}
