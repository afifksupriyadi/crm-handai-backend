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

**Period Types & Parameters:**
- **DAILY**: Menampilkan setiap hari dalam bulan
  - Required: period=DAILY&year=2025&month=12
  - Optional: week=1 (filter by specific week in month)
  
- **WEEKLY**: Menampilkan data mingguan dalam bulan
  - Required: period=WEEKLY&year=2025&month=12
  
- **MONTHLY**: Menampilkan data bulanan dalam tahun
  - Required: period=MONTHLY&year=2025
  
- **YEARLY**: Menampilkan data tahunan
  - Required: period=YEARLY&year=2025

**Forecast Calculation:**
- Normal = AVG dari 3 periode sebelumnya
- Minimum = MIN dari 3 periode sebelumnya
- Maximum = MAX dari 3 periode sebelumnya

**Important Notes:**
- Forecast membutuh data historis 3 periode sebelumnya
- Jika data historis tidak cukup, akan return error "Data historis tidak cukup"
- Forecast dengan nilai 0 (insufficient data) akan di-filter otomatis

**Response Example:**
- DAILY: Up to 30/31 days data (hanya yang punya data historis)
- WEEKLY: Up to 4-5 weeks data (hanya yang punya data historis)
- MONTHLY: Up to 12 months data (hanya yang punya data historis)
- YEARLY: 1 year data (jika ada data historis 3 tahun sebelumnya)`,
			Tags: []string{"forecasting"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetForecasts,
	)
}
