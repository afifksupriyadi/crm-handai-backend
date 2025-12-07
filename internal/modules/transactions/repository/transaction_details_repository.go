package repository

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/uptrace/bun"
)

type TransactionDetailRepository interface {
	GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error)
	GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error)
	CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error
	CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error
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
		return nil, err
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
		return nil, err
	}

	return details, nil
}

func (r *TransactionDetailRepositoryImpl) CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error {
	_, err := r.db.NewInsert().
		Model(detail).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *TransactionDetailRepositoryImpl) CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error {
	_, err := r.db.NewInsert().
		Model(&details).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
