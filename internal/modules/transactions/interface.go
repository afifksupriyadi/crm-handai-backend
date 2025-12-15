package transactions

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/uptrace/bun"
)

type TransactionService interface {
	// Regular methods (without transaction)
	GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error)
	GetOrCreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error)

	// Transaction methods (for batch import)
	GetTransactionByCodeInTx(ctx context.Context, tx *bun.Tx, code string) (*model.Transaction, error)
	CreateTransactionInTx(ctx context.Context, tx *bun.Tx, transaction *model.Transaction) error
}

type TransactionDetailService interface {
	// Regular methods (without transaction)
	GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error)
	GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error)
	CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error
	CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error

	// Transaction methods (for batch import)
	CreateTransactionDetailInTx(ctx context.Context, tx *bun.Tx, detail *model.TransactionDetail) error
}
