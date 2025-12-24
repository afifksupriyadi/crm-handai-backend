package forecasting

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting/model"
)

type SalesForecastService interface {
	GenerateForecasts(ctx context.Context, endDate time.Time, transactionBatchID int) (*model.GenerateForecastsResponse, error)
	GetForecastsByPeriod(ctx context.Context, period model.ForecastPeriod, year int) ([]*model.SalesForecastResponse, error)
}
