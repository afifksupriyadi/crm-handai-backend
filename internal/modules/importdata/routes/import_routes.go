// internal/modules/importdata/routes/import_routes.go

package routes

import (
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/handler"
	"github.com/gofiber/fiber/v2"
)

func RegisterImportRoutes(app *fiber.App, h *handler.ImportHandler) {
	basePath := fmt.Sprintf("%s/import", config.Get().BasePath)

	// POST /import/customers
	app.Post(basePath+"/customers", h.HandleImportCustomers)

	// POST /import/transactions
	app.Post(basePath+"/transactions", h.HandleImportTransactions)
}
