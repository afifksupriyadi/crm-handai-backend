package transactions

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
)

type TransactionService interface {
	GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error)
	GetOrCreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error)
}

type TransactionDetailService interface {
	GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error)
	GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error)
	CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error
	CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error
}
