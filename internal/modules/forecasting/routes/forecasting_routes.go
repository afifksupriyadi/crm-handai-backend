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
  - **Menampilkan SEMUA HARI** (termasuk yang forecast-nya 0)
  
- **WEEKLY**: Menampilkan data mingguan dalam bulan
  - Required: period=WEEKLY&year=2025&month=12
  - **Menampilkan SEMUA MINGGU** (termasuk yang forecast-nya 0)
  
- **MONTHLY**: Menampilkan data bulanan dalam tahun
  - Required: period=MONTHLY&year=2025
  - **Menampilkan SEMUA 12 BULAN (Januari - Desember)** meskipun forecast-nya 0
  
- **YEARLY**: Menampilkan data tahunan
  - Required: period=YEARLY&year=2025
  - **Menampilkan HANYA TAHUN YANG ADA DATANYA** (filter yang nilai 0)

**Forecast Calculation:**
- Normal = AVG dari 3 periode sebelumnya
- Minimum = MIN dari 3 periode sebelumnya
- Maximum = MAX dari 3 periode sebelumnya
- Jika data historis < 3 periode: nilai forecast = 0

**Important Notes:**
- **MONTHLY**: Selalu return 12 bulan (Januari-Desember), jika tidak ada data historis maka nilai forecast = 0
- **YEARLY**: Hanya return tahun yang punya data historis yang cukup (minimal 3 tahun sebelumnya)
- **DAILY & WEEKLY**: Return semua periode, termasuk yang forecast-nya 0
- Forecast dengan nilai 0 menandakan data historis tidak cukup untuk periode tersebut

**Response Example:**
- DAILY: Up to 30/31 days data (all days, including zeros)
- WEEKLY: Up to 4-5 weeks data (all weeks, including zeros)
- MONTHLY: Always 12 months data (all months, including zeros)
- YEARLY: Only years with sufficient historical data (no zeros)`,
			Tags: []string{"forecasting"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetForecasts,
	)
}
