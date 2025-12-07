// internal/modules/transactions/repository/transaction_detail_repository.go

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/uptrace/bun"
)

type TransactionDetailRepository interface {
	GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error)
	GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error)
	CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error
	CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error
	UpdateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error
}

type TransactionDetailRepositoryImpl struct {
	db *bun.DB
}

func NewTransactionDetailRepository(db *bun.DB) TransactionDetailRepository {
	return &TransactionDetailRepositoryImpl{db: db}
}

func (r *TransactionDetailRepositoryImpl) GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error) {
	detail := new(model.TransactionDetail)
	err := r.db.NewSelect().
		Model(detail).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction detail not found")
		}
		return nil, fmt.Errorf("failed to get transaction detail: %w", err)
	}

	return detail, nil
}

func (r *TransactionDetailRepositoryImpl) GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error) {
	var details []*model.TransactionDetail
	err := r.db.NewSelect().
		Model(&details).
		Where("transaction_code = ?", transactionCode).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction details: %w", err)
	}

	return details, nil
}

func (r *TransactionDetailRepositoryImpl) CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error {
	_, err := r.db.NewInsert().
		Model(detail).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create transaction detail: %w", err)
	}

	return nil
}

func (r *TransactionDetailRepositoryImpl) CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error {
	_, err := r.db.NewInsert().
		Model(&details).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create transaction details bulk: %w", err)
	}

	return nil
}

func (r *TransactionDetailRepositoryImpl) UpdateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error {
	_, err := r.db.NewUpdate().
		Model(detail).
		Where("id = ?", detail.ID).
		Where("deleted_at IS NULL").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update transaction detail: %w", err)
	}

	return nil
}
