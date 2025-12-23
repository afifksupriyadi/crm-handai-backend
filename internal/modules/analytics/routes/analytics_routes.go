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

	// GET /api/analytics/customers/loyal
	huma.Register(api,
		huma.Operation{
			OperationID: "get-loyal-customers",
			Method:      http.MethodGet,
			Path:        basePath + "/customers/loyal",
			Summary:     "Get Loyal Customers",
			Description: "Retrieve list of loyal customers (3 consecutive correct predictions).",
			Tags:        []string{"analytics"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetLoyalCustomers,
	)

	// GET /api/analytics/customers/churn
	huma.Register(api,
		huma.Operation{
			OperationID: "get-churn-customers",
			Method:      http.MethodGet,
			Path:        basePath + "/customers/churn",
			Summary:     "Get Churn Customers",
			Description: "Retrieve list of churn (at-risk) customers (2 consecutive false predictions).",
			Tags:        []string{"analytics"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetChurnCustomers,
	)
}
