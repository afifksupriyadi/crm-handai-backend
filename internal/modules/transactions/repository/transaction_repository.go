// internal/modules/transactions/repository/transaction_repository.go

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/uptrace/bun"
)

type TransactionRepository interface {
	GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error)
	CreateTransaction(ctx context.Context, transaction *model.Transaction) error
	GetOrCreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *model.Transaction) error
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return transaction, nil
}

func (r *TransactionRepositoryImpl) CreateTransaction(ctx context.Context, transaction *model.Transaction) error {
	_, err := r.db.NewInsert().
		Model(transaction).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepositoryImpl) GetOrCreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error) {
	existing, err := r.GetTransactionByCode(ctx, transaction.Code)
	if err == nil {
		return existing, nil
	}

	err = r.CreateTransaction(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return transaction, nil
}

func (r *TransactionRepositoryImpl) UpdateTransaction(ctx context.Context, transaction *model.Transaction) error {
	_, err := r.db.NewUpdate().
		Model(transaction).
		Where("code = ?", transaction.Code).
		Where("deleted_at IS NULL").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}
