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
  - **Forecast Logic**: Forecast hari X = AVG dari 3 HARI SEBELUMNYA (X-1, X-2, X-3)
  - Contoh: Forecast 15 Des = AVG(14 Des, 13 Des, 12 Des)
  
- **WEEKLY**: Menampilkan data mingguan dalam bulan
  - Required: period=WEEKLY&year=2025&month=12
  - **Forecast Logic**: Forecast minggu X = AVG dari 3 MINGGU SEBELUMNYA
  - Contoh: Forecast Week 4 = AVG(Week 3, Week 2, Week 1)
  
- **MONTHLY**: Menampilkan data bulanan dalam tahun
  - Required: period=MONTHLY&year=2025
  - **Forecast Logic**: Forecast bulan X = AVG dari 3 BULAN SEBELUMNYA
  - Contoh: Forecast Desember = AVG(November, Oktober, September)
  - **Menampilkan SEMUA 12 BULAN** meskipun forecast-nya 0
  
- **YEARLY**: Menampilkan data tahunan
  - Required: period=YEARLY&year=2025
  - **Forecast Logic**: Forecast tahun X = AVG dari 3 TAHUN SEBELUMNYA
  - Contoh: Forecast 2026 = AVG(2025, 2024, 2023)
  - **Menampilkan HANYA TAHUN YANG ADA DATANYA**

**Forecast Calculation Formula:**
- Normal (Prediksi Normal) = AVG dari periode sebelumnya
- Minimum (Prediksi Minimum) = MIN dari periode sebelumnya
- Maximum (Prediksi Maximum) = MAX dari periode sebelumnya

**Important Notes:**
- **Rolling Window**: Semua forecast menggunakan data SEQUENTIAL (periode sebelumnya berturut-turut)
- **MONTHLY**: Selalu return 12 bulan, jika tidak ada 3 bulan sebelumnya maka nilai = 0
- **YEARLY**: Hanya return tahun yang punya data 3 tahun sebelumnya
- **DAILY & WEEKLY**: Return semua periode, termasuk yang forecast-nya 0
- Minimal data yang dibutuhkan: 3 periode sebelumnya untuk generate forecast

**Response Example:**
- DAILY: Up to 30/31 days (hari ke-4 dst baru ada forecast jika import sequential)
- WEEKLY: Up to 4-5 weeks (minggu ke-4 dst baru ada forecast jika import sequential)
- MONTHLY: Always 12 months (bulan ke-4 dst baru ada forecast jika import sequential)
- YEARLY: Only years with 3 years historical data`,
			Tags: []string{"forecasting"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetForecasts,
	)
}
