package routes

import (
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterImportRoutes registers import endpoints using native Fiber
// Note: These endpoints use multipart/form-data for file uploads which is not well supported by Huma OpenAPI
// They won't appear in the OpenAPI docs but are fully functional
func RegisterImportRoutes(app *fiber.App, h *handler.ImportHandler) {
	basePath := fmt.Sprintf("%s/import", config.Get().BasePath)

	// POST /api/import/batch - Import both customer and transaction files as a batch
	// Form fields: file_customer (optional, .xlsx), file_transaction (required, .xlsx), batch_date (required, YYYY-MM-DD), notes (optional)
	app.Post(basePath+"/batch", h.HandleImportBatch)
}
