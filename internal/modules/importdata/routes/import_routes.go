package routes

import (
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/handler"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
)

// RegisterImportRoutes registers import endpoints
// Uses both Huma (for docs) and native Fiber (for actual handling)
func RegisterImportRoutes(api huma.API, app *fiber.App, h *handler.ImportHandler) {
	basePath := fmt.Sprintf("%s/import", config.Get().BasePath)

	// Register manual OpenAPI schema for documentation
	registerImportBatchDocs(api, basePath)

	// Actual implementation using native Fiber (for multipart support)
	app.Post(basePath+"/batch", h.HandleImportBatch)
}

// registerImportBatchDocs adds manual OpenAPI documentation for import batch endpoint
func registerImportBatchDocs(api huma.API, basePath string) {
	// Get OpenAPI instance
	openapi := api.OpenAPI()

	// Add to OpenAPI paths
	if openapi.Paths == nil {
		openapi.Paths = make(map[string]*huma.PathItem)
	}

	path := basePath + "/batch"
	if openapi.Paths[path] == nil {
		openapi.Paths[path] = &huma.PathItem{}
	}

	// Define the operation manually
	openapi.Paths[path].Post = &huma.Operation{
		OperationID: "import-batch",
		Summary:     "Import Customer and Transaction Batch",
		Description: `Import customer and transaction data from Excel files as a batch.

**Important Notes:**
- This endpoint accepts multipart/form-data
- Transaction file (file_transaction) is REQUIRED
- Customer file (file_customer) is OPTIONAL
- Files must follow naming convention: Transaksi_Kasir_Warung_DDMMYY_HHMMSS.xlsx

**Example cURL:**
` + "```bash" + `
curl -X POST "http://localhost:8080/api/import/batch" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file_transaction=@Transaksi_Kasir_Warung_161025_143434.xlsx" \
  -F "file_customer=@Transaksi_Pelanggan_Kasir_Warung_201025_213949.xlsx" \
  -F "start_date=2025-10-16" \
  -F "end_date=2025-10-20" \
  -F "notes=Batch import October 2025"
` + "```",
		Tags: []string{"import"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		RequestBody: &huma.RequestBody{
			Required:    true,
			Description: "Multipart form data containing Excel files and parameters",
			Content: map[string]*huma.MediaType{
				"multipart/form-data": {
					Schema: &huma.Schema{
						Type: "object",
						Properties: map[string]*huma.Schema{
							"file_customer": {
								Type:        "string",
								Format:      "binary",
								Description: "Customer Excel file (optional) - Format: Transaksi_Pelanggan_Kasir_Warung_DDMMYY_HHMMSS.xlsx",
							},
							"file_transaction": {
								Type:        "string",
								Format:      "binary",
								Description: "Transaction Excel file (required) - Format: Transaksi_Kasir_Warung_DDMMYY_HHMMSS.xlsx",
							},
							"start_date": {
								Type:        "string",
								Format:      "date",
								Description: "Start date of import range (YYYY-MM-DD)",
							},
							"end_date": {
								Type:        "string",
								Format:      "date",
								Description: "End date of import range (YYYY-MM-DD)",
							},
							"notes": {
								Type:        "string",
								Description: "Optional notes for this batch import",
							},
						},
						Required: []string{"file_transaction", "start_date", "end_date"},
					},
				},
			},
		},
		Responses: map[string]*huma.Response{
			"200": {
				Description: "Batch import successful",
				Content: map[string]*huma.MediaType{
					"application/json": {
						Schema: &huma.Schema{
							Type: "object",
							Properties: map[string]*huma.Schema{
								"data": {
									Type: "object",
									Properties: map[string]*huma.Schema{
										"batch": {
											Type: "object",
											Properties: map[string]*huma.Schema{
												"id":         {Type: "integer"},
												"batch_code": {Type: "string"},
												"batch_date": {Type: "string"},
												"status":     {Type: "string"},
												"is_active":  {Type: "boolean"},
											},
										},
										"customers": {
											Type: "object",
											Properties: map[string]*huma.Schema{
												"totalRows":        {Type: "integer"},
												"successRows":      {Type: "integer"},
												"failedRows":       {Type: "integer"},
												"customersCreated": {Type: "integer"},
												"customersUpdated": {Type: "integer"},
											},
										},
										"transactions": {
											Type: "object",
											Properties: map[string]*huma.Schema{
												"totalRows":                 {Type: "integer"},
												"successRows":               {Type: "integer"},
												"failedRows":                {Type: "integer"},
												"transactionsCreated":       {Type: "integer"},
												"transactionDetailsCreated": {Type: "integer"},
											},
										},
									},
								},
								"code":    {Type: "string"},
								"message": {Type: "string"},
							},
						},
					},
				},
			},
			"400": {
				Description: "Bad request - Invalid input",
				Content: map[string]*huma.MediaType{
					"application/json": {
						Schema: &huma.Schema{
							Type: "object",
							Properties: map[string]*huma.Schema{
								"code":    {Type: "string"},
								"message": {Type: "string"},
								"data":    {Type: "object"},
							},
						},
					},
				},
			},
			"401": {
				Description: "Unauthorized - Invalid or missing token",
			},
			"500": {
				Description: "Internal server error",
			},
		},
	}
}
