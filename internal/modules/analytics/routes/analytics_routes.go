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

	// GET /api/analytics/churn-customers
	huma.Register(api,
		huma.Operation{
			OperationID: "get-churn-customers",
			Method:      http.MethodGet,
			Path:        basePath + "/churn-customers",
			Summary:     "Get Possible Churn Customers",
			Description: "Retrieve paginated list of customers with high churn risk based on purchase behavior.",
			Tags:        []string{"analytics"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetChurnCustomers,
	)

	// GET /api/analytics/loyal-customers
	huma.Register(api,
		huma.Operation{
			OperationID: "get-loyal-customers",
			Method:      http.MethodGet,
			Path:        basePath + "/loyal-customers",
			Summary:     "Get Loyal Customers",
			Description: "Retrieve paginated list of loyal customers with their recent purchase statistics.",
			Tags:        []string{"analytics"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetLoyalCustomers,
	)
}
