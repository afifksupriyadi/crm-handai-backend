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
  - Optional: period=YEARLY (return semua tahun)
  - Optional: period=YEARLY&year=2025 (return tahun spesifik)
  - **Forecast Logic**: Forecast tahun X = AVG dari 3 TAHUN SEBELUMNYA
  - Contoh: Forecast 2026 = AVG(2025, 2024, 2023)
  - **Generate forecast sampai +1 tahun ke depan**
  - **Menampilkan HANYA TAHUN YANG ADA DATANYA (nilai non-zero)**

**Forecast Calculation Formula:**
- Normal (Prediksi Normal) = AVG dari periode sebelumnya
- Minimum (Prediksi Minimum) = MIN dari periode sebelumnya
- Maximum (Prediksi Maximum) = MAX dari periode sebelumnya

**Important Notes - YEARLY:**
- **Automatic Future Forecast**: Setiap import akan generate forecast +1 tahun ke depan
- **Data 2022-2025** → Generate forecast untuk **2025 dan 2026**
  - Forecast 2025 = AVG(2024, 2023, 2022)
  - Forecast 2026 = AVG(2025, 2024, 2023)
- **Minimal Data**: Butuh 3 tahun sebelumnya untuk generate forecast
- **Return All Years**: Jika tidak kasih parameter year, return SEMUA tahun yang ada forecast

**Important Notes - OTHER PERIODS:**
- **Full Range Generation**: Setiap import akan regenerate forecast untuk SEMUA periode dari awal data sampai sekarang
- **DAILY**: Generate untuk SEMUA hari dari bulan pertama yang ada data sampai bulan import terakhir
- **WEEKLY**: Generate untuk SEMUA minggu dari bulan pertama yang ada data sampai bulan import terakhir
- **MONTHLY**: Selalu return 12 bulan, jika tidak ada 3 bulan sebelumnya maka nilai = 0

**API Usage Examples:**

1. **Get ALL yearly forecasts**:
   ` + "`GET /api/forecasting/sales?period=YEARLY`" + `
   Response: [2025, 2026] (semua tahun yang ada data)

2. **Get specific year**:
   ` + "`GET /api/forecasting/sales?period=YEARLY&year=2026`" + `
   Response: [2026] (tahun 2026 saja)

3. **Get monthly for specific year**:
   ` + "`GET /api/forecasting/sales?period=MONTHLY&year=2025`" + `
   Response: 12 months (Jan-Dec 2025)

4. **Get daily for specific month**:
   ` + "`GET /api/forecasting/sales?period=DAILY&year=2025&month=12`" + `
   Response: All days in December 2025

**Example Scenario (Data 2022-2025):**

Import data Sept-Dec 2025 akan generate:
- **DAILY**: ~120 hari (Sept-Dec, semua hari)
- **WEEKLY**: ~17 minggu (Sept-Dec, semua minggu)
- **MONTHLY**: 12 bulan × 4 tahun = 48 records (2022-2025, all months)
- **YEARLY**: 2 records (2025, 2026)
  - 2025 = AVG(2024, 2023, 2022)
  - 2026 = AVG(2025, 2024, 2023) ✅ Future forecast!`,
			Tags: []string{"forecasting"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetForecasts,
	)
}
