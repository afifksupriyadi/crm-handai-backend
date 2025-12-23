package importdata

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
)

type ImportService interface {
	ImportBatch(ctx context.Context, customerFile, transactionFile multipart.File, customerFilename, transactionFilename, notes string, startDate, endDate time.Time) (*model.ImportBatchResponse, error) // ← ADD 2 PARAMS

}
