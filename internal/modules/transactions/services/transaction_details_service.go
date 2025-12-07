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

type TransactionDetailServiceImpl struct {
	detailRepo repository.TransactionDetailRepository
}

func NewTransactionDetailService(detailRepo repository.TransactionDetailRepository) transactions.TransactionDetailService {
	return &TransactionDetailServiceImpl{
		detailRepo: detailRepo,
	}
}

func (s *TransactionDetailServiceImpl) GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error) {
	detail, err := s.detailRepo.GetTransactionDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrTransactionDetailsNotFound, "Transaction detail not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get transaction detail")
	}

	return detail, nil
}

func (s *TransactionDetailServiceImpl) GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error) {
	details, err := s.detailRepo.GetTransactionDetailsByTransactionCode(ctx, transactionCode)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get transaction details")
	}

	return details, nil
}

func (s *TransactionDetailServiceImpl) CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error {
	err := s.detailRepo.CreateTransactionDetail(ctx, detail)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create transaction detail")
	}

	return nil
}

func (s *TransactionDetailServiceImpl) CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error {
	err := s.detailRepo.CreateTransactionDetailsBulk(ctx, details)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create transaction details bulk")
	}

	return nil
}
