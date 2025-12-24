package routes

import (
	"fmt"
	"net/http"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting/handler"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterForecastingRoutes(api huma.API, h *handler.ForecastingHandler) {
	basePath := fmt.Sprintf("%s/forecasting", config.Get().BasePath)

	huma.Register(api,
		huma.Operation{
			OperationID: "get-sales-forecasts",
			Method:      http.MethodGet,
			Path:        basePath + "/sales",
			Summary:     "Get Sales Forecasts",
			Description: `Retrieve sales forecasts by period and year.

**Request Parameters:**
- WEEKLY: period=WEEKLY&year=2025&month=10&week=1
- MONTHLY: period=MONTHLY&year=2025&month=10
- YEARLY: period=YEARLY&year=2025

**Response will contain:**
- WEEKLY: 7 days (Mon-Sun) for the specified week
- MONTHLY: 4-5 weeks for the specified month
- YEARLY: 12 months (Jan-Dec) for the specified year`,
			Tags: []string{"forecasting"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetForecasts,
	)
}
