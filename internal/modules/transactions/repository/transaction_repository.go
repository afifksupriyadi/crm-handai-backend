package repository

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/uptrace/bun"
)

type TransactionRepository interface {
	GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error)
	CreateTransaction(ctx context.Context, transaction *model.Transaction) error
}

type TransactionRepositoryImpl struct {
	db *bun.DB
}

func NewTransactionRepository(db *bun.DB) TransactionRepository {
	return &TransactionRepositoryImpl{db: db}
}

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
