package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/handler"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/request"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
)

// RegisterImportRoutes registers import routes with Huma API
func RegisterImportRoutes(api huma.API, app *fiber.App, h *handler.ImportHandler) {
	basePath := fmt.Sprintf("%s/import", config.Get().BasePath)

	// POST /api/import/batch - Import batch with transaction and optional customer file
	huma.Register(api,
		huma.Operation{
			OperationID: "importBatch",
			Method:      http.MethodPost,
			Path:        basePath + "/batch",
			Summary:     "Import Batch",
			Description: "Import transaction data (required) and optionally customer data as a batch. **Note:** This endpoint requires multipart/form-data. Use file_transaction (required), file_customer (optional), batch_date (YYYY-MM-DD), overwrite_if_exist (true/false), notes (optional).",
			Tags:        []string{"import"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			MaxBodyBytes: 50 * 1024 * 1024, // 50MB max
		}, func(ctx context.Context, input *request.GenericRequest[model.ImportBatchDocRequest]) (*response.Response, error) {
			return response.BuildSuccess(map[string]interface{}{
				"note":               "Use multipart/form-data",
				"file_transaction":   "Excel file (.xlsx) - Required",
				"file_customer":      "Excel file (.xlsx) - Optional",
				"batch_date":         "YYYY-MM-DD format - Required",
				"overwrite_if_exist": "true/false - Optional (default: false)",
				"notes":              "Text - Optional",
			}, response.SuccessImportBatch), nil
		},
	)

	// POST /api/import/customers - Legacy customer import
	huma.Register(api,
		huma.Operation{
			OperationID: "importCustomers",
			Method:      http.MethodPost,
			Path:        basePath + "/customers",
			Summary:     "Import Customers (Legacy)",
			Description: "Legacy endpoint: Import customer data from Excel file. **Note:** This endpoint requires multipart/form-data with field 'file' (.xlsx format).",
			Tags:        []string{"import"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			MaxBodyBytes: 20 * 1024 * 1024, // 20MB max
		}, func(ctx context.Context, input *request.GenericRequest[model.ImportFileDocRequest]) (*response.Response, error) {
			return response.BuildSuccess(map[string]interface{}{
				"note": "Use multipart/form-data with field 'file' (.xlsx)",
			}, response.SuccessImportCustomers), nil
		},
	)

	// POST /api/import/transactions - Legacy transaction import
	huma.Register(api,
		huma.Operation{
			OperationID: "importTransactions",
			Method:      http.MethodPost,
			Path:        basePath + "/transactions",
			Summary:     "Import Transactions (Legacy)",
			Description: "Legacy endpoint: Import transaction data from Excel file. **Note:** This endpoint requires multipart/form-data with field 'file' (.xlsx format).",
			Tags:        []string{"import"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			MaxBodyBytes: 20 * 1024 * 1024, // 20MB max
		}, func(ctx context.Context, input *request.GenericRequest[model.ImportFileDocRequest]) (*response.Response, error) {
			return response.BuildSuccess(map[string]interface{}{
				"note": "Use multipart/form-data with field 'file' (.xlsx)",
			}, response.SuccessImportTransactions), nil
		},
	)

	// Actual Fiber handlers for file upload (multipart/form-data)
	// These override the Huma handlers above for actual implementation
	app.Post(basePath+"/batch", h.HandleImportBatch)
	app.Post(basePath+"/customers", h.HandleImportCustomers)
	app.Post(basePath+"/transactions", h.HandleImportTransactions)
}
