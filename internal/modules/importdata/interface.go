package importdata

import (
	"context"
	"mime/multipart"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
)

type ImportService interface {
	ImportCustomers(ctx context.Context, file multipart.File, filename string) (*model.ImportCustomerResponse, error)
	ImportTransactions(ctx context.Context, file multipart.File, filename string) (*model.ImportTransactionResponse, error)
}
