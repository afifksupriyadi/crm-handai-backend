package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

type TransactionServiceImpl struct {
	transactionRepo repository.TransactionRepository
}

func NewTransactionService(transactionRepo repository.TransactionRepository) transactions.TransactionService {
	return &TransactionServiceImpl{
		transactionRepo: transactionRepo,
	}
}

func (s *TransactionServiceImpl) GetTransactionByCode(ctx context.Context, code string) (*model.Transaction, error) {
	transaction, err := s.transactionRepo.GetTransactionByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrTransactionNotFound, "Transaction not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get transaction")
	}

	return transaction, nil
}

func (s *TransactionServiceImpl) GetOrCreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error) {
	// Try to get existing transaction
	existing, err := s.transactionRepo.GetTransactionByCode(ctx, transaction.Code)

	// If found, return it
	if err == nil {
		return existing, nil
	}

	// If not found, create new
	if errors.Is(err, sql.ErrNoRows) {
		err = s.transactionRepo.CreateTransaction(ctx, transaction)
		if err != nil {
			return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create transaction")
		}
		return transaction, nil
	}

	// Other database error
	return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check existing transaction")
}
