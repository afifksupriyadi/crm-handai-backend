package routes

import (
	"net/http"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/health/handler"
	"github.com/danielgtaylor/huma/v2"
)

// RegisterHealthRoutes registers the health check and welcome routes.
func RegisterHealthRoutes(api huma.API, h *handler.HealthHandler) {
	// GET / - Welcome endpoint
	huma.Register(api,
		huma.Operation{
			OperationID: "welcome",
			Method:      http.MethodGet,
			Path:        "/",
			Summary:     "Welcome Message",
			Description: "Returns a welcome message for the API",
			Tags:        []string{"health"},
			Middlewares: huma.Middlewares{},
		}, h.HandleWelcome,
	)

	// GET /health - Health check endpoint
	huma.Register(api,
		huma.Operation{
			OperationID: "healthCheck",
			Method:      http.MethodGet,
			Path:        "/health",
			Summary:     "Health Check",
			Description: "Returns the health status and uptime of the API",
			Tags:        []string{"health"},
			Middlewares: huma.Middlewares{},
		}, h.HandleHealthCheck,
	)
}
