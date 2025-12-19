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

type TransactionDetailServiceImpl struct {
	transactionDetailRepo repository.TransactionDetailRepository // ✅ ONLY repo!
}

func NewTransactionDetailService(transactionDetailRepo repository.TransactionDetailRepository) transactions.TransactionDetailService {
	return &TransactionDetailServiceImpl{
		transactionDetailRepo: transactionDetailRepo,
	}
}

func (s *TransactionDetailServiceImpl) GetTransactionDetailByID(ctx context.Context, id int) (*model.TransactionDetail, error) {
	detail, err := s.transactionDetailRepo.GetTransactionDetailByID(ctx, id)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get transaction detail")
	}
	if detail == nil {
		return nil, response.WrapAppError(ctx, nil, response.ErrTransactionDetailsNotFound, "Transaction detail not found")
	}
	return detail, nil
}

func (s *TransactionDetailServiceImpl) GetTransactionDetailsByTransactionCode(ctx context.Context, transactionCode string) ([]*model.TransactionDetail, error) {
	details, err := s.transactionDetailRepo.GetTransactionDetailsByTransactionCode(ctx, transactionCode)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get transaction details")
	}
	return details, nil
}

func (s *TransactionDetailServiceImpl) CreateTransactionDetail(ctx context.Context, detail *model.TransactionDetail) error {
	err := s.transactionDetailRepo.CreateTransactionDetail(ctx, detail)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create transaction detail")
	}

	logger.FromContext(ctx, 1).Info().Str("transaction_code", detail.TransactionCode).Int("product_id", detail.ProductID).Msg("Transaction detail created")
	return nil
}

func (s *TransactionDetailServiceImpl) CreateTransactionDetailInTx(ctx context.Context, tx *bun.Tx, detail *model.TransactionDetail) error {
	err := s.transactionDetailRepo.CreateTransactionDetailInTx(ctx, tx, detail)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create transaction detail")
	}

	logger.FromContext(ctx, 1).Debug().Str("transaction_code", detail.TransactionCode).Int("product_id", detail.ProductID).Msg("Transaction detail created in batch")
	return nil
}

func (s *TransactionDetailServiceImpl) CreateTransactionDetailsBulk(ctx context.Context, details []*model.TransactionDetail) error {
	err := s.transactionDetailRepo.CreateTransactionDetailsBulk(ctx, details)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create transaction details")
	}
	return nil
}
