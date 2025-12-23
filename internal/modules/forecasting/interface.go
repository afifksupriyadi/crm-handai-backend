package forecasting

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting/model"
)

// SalesForecastService defines the contract for forecasting operations
type SalesForecastService interface {
	// GenerateForecasts creates forecasts after import completion
	GenerateForecasts(ctx context.Context, endDate time.Time, transactionBatchID int) (*model.GenerateForecastsResponse, error)

	// GetForecastsByPeriod retrieves forecasts for display
	GetForecastsByPeriod(ctx context.Context, period model.ForecastPeriod, year int) ([]*model.SalesForecastResponse, error)
}
