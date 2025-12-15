package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/handler"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
)

// RegisterImportRoutes registers import routes with both Huma (for docs) and Fiber (for file upload)
func RegisterImportRoutes(api huma.API, app *fiber.App, h *handler.ImportHandler) {
	basePath := fmt.Sprintf("%s/import", config.Get().BasePath)

	// Register with Huma for OpenAPI documentation
	// Note: Actual handlers use Fiber because Huma v2 doesn't support multipart well

	// POST /api/import/batch - Import batch with transaction and optional customer file
	huma.Register(api,
		huma.Operation{
			OperationID: "importBatch",
			Method:      http.MethodPost,
			Path:        basePath + "/batch",
			Summary:     "Import Batch",
			Description: "Import transaction data (required) and optionally customer data as a batch. Use multipart/form-data with fields: file_transaction (required), file_customer (optional), batch_date (YYYY-MM-DD), overwrite_if_exist (true/false), notes (optional).",
			Tags:        []string{"import"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, func(ctx context.Context, input *struct{}) (*struct{ Body any }, error) {
			// Dummy handler for docs only - actual implementation in Fiber
			return &struct{ Body any }{Body: map[string]string{"message": "Use multipart/form-data with Fiber endpoint"}}, nil
		},
	)

	// POST /api/import/customers - Legacy customer import
	huma.Register(api,
		huma.Operation{
			OperationID: "importCustomers",
			Method:      http.MethodPost,
			Path:        basePath + "/customers",
			Summary:     "Import Customers (Legacy)",
			Description: "Legacy endpoint: Import customer data from Excel file. Use multipart/form-data with field 'file'.",
			Tags:        []string{"import"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, func(ctx context.Context, input *struct{}) (*struct{ Body any }, error) {
			return &struct{ Body any }{Body: map[string]string{"message": "Use multipart/form-data with Fiber endpoint"}}, nil
		},
	)

	// POST /api/import/transactions - Legacy transaction import
	huma.Register(api,
		huma.Operation{
			OperationID: "importTransactions",
			Method:      http.MethodPost,
			Path:        basePath + "/transactions",
			Summary:     "Import Transactions (Legacy)",
			Description: "Legacy endpoint: Import transaction data from Excel file. Use multipart/form-data with field 'file'.",
			Tags:        []string{"import"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, func(ctx context.Context, input *struct{}) (*struct{ Body any }, error) {
			return &struct{ Body any }{Body: map[string]string{"message": "Use multipart/form-data with Fiber endpoint"}}, nil
		},
	)

	// Actual Fiber handlers for file upload
	app.Post(basePath+"/batch", h.HandleImportBatch)
	app.Post(basePath+"/customers", h.HandleImportCustomers)
	app.Post(basePath+"/transactions", h.HandleImportTransactions)
}
