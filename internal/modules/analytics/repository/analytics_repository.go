package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/model"
	"github.com/uptrace/bun"
)

type AnalyticsRepository interface {
	GetSalesDataByDaily(ctx context.Context, startDate, endDate time.Time) ([]*model.SalesDataPoint, error)
	GetSalesDataByMonthly(ctx context.Context, startDate, endDate time.Time) ([]*model.SalesDataPoint, error)
	GetSalesDataByYearly(ctx context.Context, startDate, endDate time.Time) ([]*model.SalesDataPoint, error)
}

type AnalyticsRepositoryImpl struct {
	db *bun.DB
}

func NewAnalyticsRepository(db *bun.DB) AnalyticsRepository {
	return &AnalyticsRepositoryImpl{db: db}
}

func (r *AnalyticsRepositoryImpl) GetSalesDataByDaily(ctx context.Context, startDate, endDate time.Time) ([]*model.SalesDataPoint, error) {
	var results []struct {
		Date    time.Time `bun:"date"`
		Revenue float64   `bun:"revenue"`
	}

	query := `
		WITH transaction_totals AS (
			SELECT 
				t.code,
				t.transaction_date::date as date,
				COALESCE(SUM(td.subtotal), 0) as items_total,
				t.discount,
				t.shipping_cost
			FROM transactions t
			LEFT JOIN transaction_details td ON t.code = td.transaction_code AND td.deleted_at IS NULL
			WHERE 
				t.transaction_date >= ? 
				AND t.transaction_date < ?
				AND t.status = 'LUNAS'
				AND t.deleted_at IS NULL
			GROUP BY t.code, date, t.discount, t.shipping_cost
		)
		SELECT 
			date,
			COALESCE(SUM(items_total - discount + shipping_cost), 0) as revenue
		FROM transaction_totals
		GROUP BY date
		ORDER BY date
	`

	err := r.db.NewRaw(query, startDate, endDate.AddDate(0, 0, 1)).Scan(ctx, &results)
	if err != nil {
		return nil, err
	}

	dataPoints := make([]*model.SalesDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, &model.SalesDataPoint{
			Period:      result.Date.Format("02 Jan"),
			PeriodOrder: result.Date.YearDay(),
			Date:        result.Date,
			Revenue:     result.Revenue,
		})
	}

	return dataPoints, nil
}

func (r *AnalyticsRepositoryImpl) GetSalesDataByMonthly(ctx context.Context, startDate, endDate time.Time) ([]*model.SalesDataPoint, error) {
	var results []struct {
		Year    int     `bun:"year"`
		Month   int     `bun:"month"`
		Revenue float64 `bun:"revenue"`
	}

	query := `
		WITH transaction_totals AS (
			SELECT 
				t.code,
				EXTRACT(YEAR FROM t.transaction_date)::int as year,
				EXTRACT(MONTH FROM t.transaction_date)::int as month,
				COALESCE(SUM(td.subtotal), 0) as items_total,
				t.discount,
				t.shipping_cost
			FROM transactions t
			LEFT JOIN transaction_details td ON t.code = td.transaction_code AND td.deleted_at IS NULL
			WHERE 
				t.transaction_date >= ? 
				AND t.transaction_date < ?
				AND t.status = 'LUNAS'
				AND t.deleted_at IS NULL
			GROUP BY t.code, year, month, t.discount, t.shipping_cost
		)
		SELECT 
			year,
			month,
			COALESCE(SUM(items_total - discount + shipping_cost), 0) as revenue
		FROM transaction_totals
		GROUP BY year, month
		ORDER BY year, month
	`

	err := r.db.NewRaw(query, startDate, endDate.AddDate(0, 0, 1)).Scan(ctx, &results)
	if err != nil {
		return nil, err
	}

	dataPoints := make([]*model.SalesDataPoint, 0, len(results))
	monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for _, result := range results {
		date := time.Date(result.Year, time.Month(result.Month), 1, 0, 0, 0, 0, time.UTC)

		dataPoints = append(dataPoints, &model.SalesDataPoint{
			Period:      monthNames[result.Month],
			PeriodOrder: result.Month,
			Date:        date,
			Revenue:     result.Revenue,
		})
	}

	return dataPoints, nil
}

func (r *AnalyticsRepositoryImpl) GetSalesDataByYearly(ctx context.Context, startDate, endDate time.Time) ([]*model.SalesDataPoint, error) {
	var results []struct {
		Year    int     `bun:"year"`
		Revenue float64 `bun:"revenue"`
	}

	query := `
		WITH transaction_totals AS (
			SELECT 
				t.code,
				EXTRACT(YEAR FROM t.transaction_date)::int as year,
				COALESCE(SUM(td.subtotal), 0) as items_total,
				t.discount,
				t.shipping_cost
			FROM transactions t
			LEFT JOIN transaction_details td ON t.code = td.transaction_code AND td.deleted_at IS NULL
			WHERE 
				t.transaction_date >= ? 
				AND t.transaction_date < ?
				AND t.status = 'LUNAS'
				AND t.deleted_at IS NULL
			GROUP BY t.code, year, t.discount, t.shipping_cost
		)
		SELECT 
			year,
			COALESCE(SUM(items_total - discount + shipping_cost), 0) as revenue
		FROM transaction_totals
		GROUP BY year
		ORDER BY year
	`

	err := r.db.NewRaw(query, startDate, endDate.AddDate(0, 0, 1)).Scan(ctx, &results)
	if err != nil {
		return nil, err
	}

	dataPoints := make([]*model.SalesDataPoint, 0, len(results))
	for _, result := range results {
		date := time.Date(result.Year, 1, 1, 0, 0, 0, 0, time.UTC)

		dataPoints = append(dataPoints, &model.SalesDataPoint{
			Period:      fmt.Sprintf("%d", result.Year),
			PeriodOrder: result.Year,
			Date:        date,
			Revenue:     result.Revenue,
		})
	}

	return dataPoints, nil
}
