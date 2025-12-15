package repository

import (
	"context"
	"database/sql"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/uptrace/bun"
)

type TransactionRepository interface {
	// Regular methods (without transaction)
	GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error)
	CreateTransaction(ctx context.Context, transaction *model.Transaction) error

	// Transaction methods (for batch import)
	GetTransactionByCodeInTx(ctx context.Context, tx *bun.Tx, code string) (*model.Transaction, error)
	CreateTransactionInTx(ctx context.Context, tx *bun.Tx, transaction *model.Transaction) error
}

type TransactionRepositoryImpl struct {
	db *bun.DB
}

func NewTransactionRepository(db *bun.DB) TransactionRepository {
	return &TransactionRepositoryImpl{db: db}
}

// ==========================================
// REGULAR METHODS (Without Transaction)
// ==========================================

func (r *TransactionRepositoryImpl) GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error) {
	transaction := new(model.Transaction)
	err := r.db.NewSelect().
		Model(transaction).
		Where("code = ?", code).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *TransactionRepositoryImpl) CreateTransaction(ctx context.Context, transaction *model.Transaction) error {
	_, err := r.db.NewInsert().
		Model(transaction).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ==========================================
// TRANSACTION METHODS (With Transaction)
// ==========================================

func (r *TransactionRepositoryImpl) GetTransactionByCodeInTx(ctx context.Context, tx *bun.Tx, code string) (*model.Transaction, error) {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	transaction := new(model.Transaction)
	err := db.NewSelect().
		Model(transaction).
		Where("code = ?", code).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *TransactionRepositoryImpl) CreateTransactionInTx(ctx context.Context, tx *bun.Tx, transaction *model.Transaction) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	_, err := db.NewInsert().
		Model(transaction).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
