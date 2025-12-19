package transactions

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/uptrace/bun"
)

// TransactionService defines the contract for transaction-related operations
type TransactionService interface {
	GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error)
	GetTransactionByCodeInTx(ctx context.Context, tx *bun.Tx, code string) (*model.Transaction, error)
	GetOrCreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error)
	CreateTransactionInTx(ctx context.Context, tx *bun.Tx, transaction *model.Transaction) error
}

// TransactionDetailService defines the contract for transaction detail operations
type TransactionDetailService interface {
	GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error)
	GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error)
	CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error
	CreateTransactionDetailInTx(ctx context.Context, tx *bun.Tx, detail *model.TransactionDetail) error
	CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error
}
