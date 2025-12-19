package service

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

type TransactionServiceImpl struct {
	transactionRepo repository.TransactionRepository // ✅ ONLY repo!
}

func NewTransactionService(transactionRepo repository.TransactionRepository) transactions.TransactionService {
	return &TransactionServiceImpl{
		transactionRepo: transactionRepo,
	}
}

func (s *TransactionServiceImpl) GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error) {
	transaction, err := s.transactionRepo.GetTransactionByCode(ctx, code)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get transaction")
	}
	if transaction == nil {
		return nil, response.WrapAppError(ctx, nil, response.ErrTransactionNotFound, "Transaction not found")
	}
	return transaction, nil
}

func (s *TransactionServiceImpl) GetTransactionByCodeInTx(ctx context.Context, tx *bun.Tx, code string) (*model.Transaction, error) {
	transaction, err := s.transactionRepo.GetTransactionByCodeInTx(ctx, tx, code)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get transaction")
	}
	return transaction, nil
}

func (s *TransactionServiceImpl) GetOrCreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error) {
	existing, err := s.transactionRepo.GetTransactionByCode(ctx, transaction.Code)
	if err == nil {
		return existing, nil
	}

	err = s.transactionRepo.CreateTransaction(ctx, transaction)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create transaction")
	}

	logger.FromContext(ctx, 1).Info().Str("code", transaction.Code).Msg("Transaction created")
	return transaction, nil
}

func (s *TransactionServiceImpl) CreateTransactionInTx(ctx context.Context, tx *bun.Tx, transaction *model.Transaction) error {
	err := s.transactionRepo.CreateTransactionInTx(ctx, tx, transaction)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create transaction")
	}

	logger.FromContext(ctx, 1).Debug().Str("code", transaction.Code).Msg("Transaction created in batch")
	return nil
}
