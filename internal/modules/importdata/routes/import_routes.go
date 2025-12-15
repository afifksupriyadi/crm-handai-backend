package routes

import (
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/handler"
	"github.com/gofiber/fiber/v2"
)

func RegisterImportRoutes(app *fiber.App, h *handler.ImportHandler) {
	basePath := fmt.Sprintf("%s/import", config.Get().BasePath)

	// NEW: POST /import/batch - Import both customer and transaction files as a batch
	app.Post(basePath+"/batch", h.HandleImportBatch)

	// LEGACY: POST /import/customers - Single customer file import (backward compatibility)
	app.Post(basePath+"/customers", h.HandleImportCustomers)

	// LEGACY: POST /import/transactions - Single transaction file import (backward compatibility)
	app.Post(basePath+"/transactions", h.HandleImportTransactions)
}
