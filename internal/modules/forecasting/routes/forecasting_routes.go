package routes

import (
	"fmt"
	"net/http"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting/handler"
	"github.com/danielgtaylor/huma/v2"
)

// RegisterForecastingRoutes registers forecasting endpoints
func RegisterForecastingRoutes(api huma.API, h *handler.ForecastingHandler) {
	basePath := fmt.Sprintf("%s/forecasting", config.Get().BasePath)

	// GET /api/forecasting/sales
	huma.Register(api,
		huma.Operation{
			OperationID: "get-sales-forecasts",
			Method:      http.MethodGet,
			Path:        basePath + "/sales",
			Summary:     "Get Sales Forecasts",
			Description: "Retrieve sales forecasts by period (WEEKLY, MONTHLY, YEARLY) and year",
			Tags:        []string{"forecasting"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetForecasts,
	)
}
