package analytics

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/model"
)

type AnalyticsService interface {
	GetSalesChart(ctx context.Context, req *model.SalesChartRequest) (*model.SalesChartResponse, error)
	GetLoyalCustomers(ctx context.Context, limit int) (*model.LoyalCustomersResponse, error)
	GetChurnCustomers(ctx context.Context, limit int) (*model.ChurnCustomersResponse, error)
}
