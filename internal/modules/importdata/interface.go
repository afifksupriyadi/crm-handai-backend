package importdata

import (
	"context"
	"mime/multipart"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
)

type ImportService interface {
	// Legacy endpoints (keep for backward compatibility)
	ImportCustomers(ctx context.Context, file multipart.File, filename string) (*model.ImportCustomerResponse, error)
	ImportTransactions(ctx context.Context, file multipart.File, filename string) (*model.ImportTransactionResponse, error)

	// New batch endpoint
	ImportBatch(ctx context.Context, customerFile, transactionFile multipart.File, customerFilename, transactionFilename, batchDate, notes string, overwriteIfExist bool) (*model.ImportBatchResponse, error)
}
