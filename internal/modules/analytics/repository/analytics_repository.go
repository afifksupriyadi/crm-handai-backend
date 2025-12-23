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
	GetChurnCustomers(ctx context.Context, page, limit int) ([]model.ChurnCustomerItem, int, error)
	GetLoyalCustomers(ctx context.Context, page, limit int) ([]model.LoyalCustomerItem, int, error)
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

// GetChurnCustomers retrieves paginated list of churn customers
func (r *AnalyticsRepositoryImpl) GetChurnCustomers(ctx context.Context, page, limit int) ([]model.ChurnCustomerItem, int, error) {
	offset := (page - 1) * limit

	var customers []model.ChurnCustomerItem
	var totalCount int

	// Get total count
	err := r.db.NewRaw(`
		SELECT COUNT(DISTINCT cs.customer_id)
		FROM analytics.customer_segments cs
		INNER JOIN customers c ON cs.customer_id = c.id
		WHERE cs.segment = 'CHURN'
			AND c.deleted_at IS NULL
	`).Scan(ctx, &totalCount)

	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	query := `
		SELECT 
			c.id,
			c.name,
			c.phone,
			COALESCE(cm.last_transaction_date, c.created_at) as last_purchase,
			COALESCE(
				EXTRACT(DAY FROM (NOW() - cm.last_transaction_date))::int,
				EXTRACT(DAY FROM (NOW() - c.created_at))::int
			) as days_since_last_purchase
		FROM analytics.customer_segments cs
		INNER JOIN customers c ON cs.customer_id = c.id
		LEFT JOIN LATERAL (
			SELECT last_transaction_date
			FROM analytics.customer_metrics
			WHERE customer_id = c.id
			ORDER BY transaction_batch_id DESC
			LIMIT 1
		) cm ON true
		WHERE cs.segment = 'CHURN'
			AND c.deleted_at IS NULL
		ORDER BY days_since_last_purchase DESC
		LIMIT ? OFFSET ?
	`

	err = r.db.NewRaw(query, limit, offset).Scan(ctx, &customers)
	if err != nil {
		return nil, 0, err
	}

	return customers, totalCount, nil
}

// GetLoyalCustomers retrieves paginated list of loyal customers
func (r *AnalyticsRepositoryImpl) GetLoyalCustomers(ctx context.Context, page, limit int) ([]model.LoyalCustomerItem, int, error) {
	offset := (page - 1) * limit
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	// Start of current week (Monday)
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday())+1)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, loc)

	// Start of current month
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

	var customers []model.LoyalCustomerItem
	var totalCount int

	// Get total count
	err := r.db.NewRaw(`
		SELECT COUNT(DISTINCT cs.customer_id)
		FROM analytics.customer_segments cs
		INNER JOIN customers c ON cs.customer_id = c.id
		WHERE cs.segment = 'LOYAL'
			AND c.deleted_at IS NULL
	`).Scan(ctx, &totalCount)

	if err != nil {
		return nil, 0, err
	}

	// Get paginated data with week and month purchase counts
	query := `
		SELECT 
			c.id,
			c.name,
			c.phone,
			COALESCE(cm.last_transaction_date, c.created_at) as last_purchase,
			COALESCE(week_purchases.total_quantity, 0) as total_product_purchases_week,
			COALESCE(month_purchases.total_quantity, 0) as total_month_purchase
		FROM analytics.customer_segments cs
		INNER JOIN customers c ON cs.customer_id = c.id
		LEFT JOIN LATERAL (
			SELECT last_transaction_date
			FROM analytics.customer_metrics
			WHERE customer_id = c.id
			ORDER BY transaction_batch_id DESC
			LIMIT 1
		) cm ON true
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(td.quantity), 0) as total_quantity
			FROM transactions t
			INNER JOIN transaction_details td ON t.code = td.transaction_code
			WHERE t.customer_id = c.id
				AND t.transaction_date >= ?
				AND t.transaction_date < ?
				AND t.deleted_at IS NULL
				AND td.deleted_at IS NULL
		) week_purchases ON true
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(td.quantity), 0) as total_quantity
			FROM transactions t
			INNER JOIN transaction_details td ON t.code = td.transaction_code
			WHERE t.customer_id = c.id
				AND t.transaction_date >= ?
				AND t.transaction_date < ?
				AND t.deleted_at IS NULL
				AND td.deleted_at IS NULL
		) month_purchases ON true
		WHERE cs.segment = 'LOYAL'
			AND c.deleted_at IS NULL
		ORDER BY cm.last_transaction_date DESC NULLS LAST
		LIMIT ? OFFSET ?
	`

	err = r.db.NewRaw(query,
		startOfWeek, now, // week range
		startOfMonth, now, // month range
		limit, offset,
	).Scan(ctx, &customers)

	if err != nil {
		return nil, 0, err
	}

	return customers, totalCount, nil
}
