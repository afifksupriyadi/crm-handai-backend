package handler

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/request"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

type ForecastingHandler struct {
	svc forecasting.SalesForecastService
}

func NewForecastingHandler(svc forecasting.SalesForecastService) *ForecastingHandler {
	return &ForecastingHandler{
		svc: svc,
	}
}

// UPDATE: forecasting_handler.go
// HandleGetForecasts retrieves forecasts by period and year
func (h *ForecastingHandler) HandleGetForecasts(ctx context.Context, req *GetSalesForecastQueryRequest) (*response.Response, error) {
	// Validate period
	if req.Period == "" {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'period' wajib diisi (WEEKLY, MONTHLY, YEARLY)")), nil
	}

	validPeriods := map[string]bool{
		"WEEKLY":  true,
		"MONTHLY": true,
		"YEARLY":  true,
	}
	if !validPeriods[req.Period] {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'period' tidak valid. Gunakan: WEEKLY, MONTHLY, atau YEARLY")), nil
	}

	// Validate year
	if req.Year == 0 {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'year' wajib diisi (contoh: 2025)")), nil
	}
	if req.Year < 2020 || req.Year > 2100 {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'year' harus antara 2020-2100")), nil
	}

	// Convert period string to ForecastPeriod
	period := model.ForecastPeriod(req.Period)

	data, err := h.svc.GetForecastsByPeriod(ctx, period, req.Year)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessForecastRetrieved), nil
}

// ADD: Request struct for query params
type GetSalesForecastQueryRequest struct {
	request.AuthorizedRequest
	Period string `query:"period" doc:"Forecast period (WEEKLY, MONTHLY, YEARLY)"`
	Year   int    `query:"year" doc:"Forecast year (2020-2100)"`
}
