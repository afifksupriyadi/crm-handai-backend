package repository

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

type SalesForecastRepository interface {
	Create(ctx context.Context, db bun.IDB, forecast *model.SalesForecast) (*model.SalesForecast, error)
	BulkCreate(ctx context.Context, db bun.IDB, forecasts []*model.SalesForecast) error
	GetByPeriodAndYear(ctx context.Context, period model.ForecastPeriod, year int) ([]*model.SalesForecast, error)
	GetHistoricalRevenue(ctx context.Context, startDate, endDate time.Time) (float64, error)
}

type SalesForecastRepositoryImpl struct {
	db *bun.DB
}

func NewSalesForecastRepository(db *bun.DB) SalesForecastRepository {
	return &SalesForecastRepositoryImpl{db: db}
}

// Create inserts a single forecast
func (r *SalesForecastRepositoryImpl) Create(ctx context.Context, db bun.IDB, forecast *model.SalesForecast) (*model.SalesForecast, error) {
	_, err := db.NewInsert().
		Model(forecast).
		On("CONFLICT (transaction_batch_id, forecast_period, forecast_date) DO UPDATE").
		Set("minimum_revenue = EXCLUDED.minimum_revenue").
		Set("normal_revenue = EXCLUDED.normal_revenue").
		Set("maximum_revenue = EXCLUDED.maximum_revenue").
		Set("computed_at = EXCLUDED.computed_at").
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create sales forecast")
		return nil, err
	}

	return forecast, nil
}

// BulkCreate inserts multiple forecasts
func (r *SalesForecastRepositoryImpl) BulkCreate(ctx context.Context, db bun.IDB, forecasts []*model.SalesForecast) error {
	if len(forecasts) == 0 {
		return nil
	}

	_, err := db.NewInsert().
		Model(&forecasts).
		On("CONFLICT (transaction_batch_id, forecast_period, forecast_date) DO UPDATE").
		Set("minimum_revenue = EXCLUDED.minimum_revenue").
		Set("normal_revenue = EXCLUDED.normal_revenue").
		Set("maximum_revenue = EXCLUDED.maximum_revenue").
		Set("computed_at = EXCLUDED.computed_at").
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to bulk create sales forecasts")
		return err
	}

	logger.FromContext(ctx, 1).Info().Int("count", len(forecasts)).Msg("Sales forecasts bulk created")
	return nil
}

// GetByPeriodAndYear retrieves forecasts for a specific period and year
func (r *SalesForecastRepositoryImpl) GetByPeriodAndYear(ctx context.Context, period model.ForecastPeriod, year int) ([]*model.SalesForecast, error) {
	var forecasts []*model.SalesForecast

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	err := r.db.NewSelect().
		Model(&forecasts).
		Where("forecast_period = ?", period).
		Where("forecast_date >= ?", startDate).
		Where("forecast_date <= ?", endDate).
		Order("forecast_date ASC").
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get forecasts by period and year")
		return nil, err
	}

	return forecasts, nil
}

// GetHistoricalRevenue calculates total revenue between date range
func (r *SalesForecastRepositoryImpl) GetHistoricalRevenue(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	var totalRevenue float64

	query := `
		SELECT COALESCE(SUM(td.subtotal), 0) as total_revenue
		FROM transaction_details td
		JOIN transactions t ON td.transaction_code = t.code
		WHERE t.transaction_date >= ? 
		  AND t.transaction_date < ?
		  AND t.deleted_at IS NULL
		  AND td.deleted_at IS NULL
		  AND t.status = 'LUNAS'
	`

	err := r.db.NewRaw(query, startDate, endDate).Scan(ctx, &totalRevenue)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get historical revenue")
		return 0, err
	}

	return totalRevenue, nil
}
