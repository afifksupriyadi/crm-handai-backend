package routes

import (
	"fmt"
	"net/http"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/handler"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterAnalyticsRoutes(api huma.API, h *handler.AnalyticsHandler) {
	basePath := fmt.Sprintf("%s/analytics", config.Get().BasePath)

	// GET /api/analytics/sales-chart
	huma.Register(api,
		huma.Operation{
			OperationID: "get-sales-chart",
			Method:      http.MethodGet,
			Path:        basePath + "/sales-chart",
			Summary:     "Get Sales Chart Data",
			Description: "Retrieve sales chart data with various period types and date ranges for analytics dashboard.",
			Tags:        []string{"analytics"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetSalesChart,
	)
}
