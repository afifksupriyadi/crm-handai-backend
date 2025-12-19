package importdata

import (
	"context"
	"mime/multipart"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
)

type ImportService interface {
	ImportBatch(ctx context.Context, customerFile, transactionFile multipart.File, customerFilename, transactionFilename, notes string) (*model.ImportBatchResponse, error)
}
