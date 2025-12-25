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
	return &ForecastingHandler{svc: svc}
}

type GetSalesForecastQueryRequest struct {
	request.AuthorizedRequest
	Period string `query:"period" doc:"Forecast period (DAILY, WEEKLY, MONTHLY, YEARLY)" example:"MONTHLY"`
	Year   int    `query:"year" minimum:"2020" maximum:"2100" doc:"Year (2020-2100) - Required for all periods" example:"2025"`
	Month  int    `query:"month" minimum:"0" maximum:"12" doc:"Month (1-12) - Required for DAILY, WEEKLY, MONTHLY" example:"12"`
	Week   int    `query:"week" minimum:"0" maximum:"5" doc:"Week number (1-5) - Optional for DAILY" example:"0"`
}

func (h *ForecastingHandler) HandleGetForecasts(ctx context.Context, req *GetSalesForecastQueryRequest) (*response.Response, error) {
	// Validate period
	if req.Period == "" {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'period' wajib diisi")), nil
	}

	validPeriods := map[string]bool{"DAILY": true, "WEEKLY": true, "MONTHLY": true, "YEARLY": true}
	if !validPeriods[req.Period] {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'period' tidak valid (harus DAILY, WEEKLY, MONTHLY, atau YEARLY)")), nil
	}

	// ✅ FIX: Year WAJIB untuk SEMUA period (termasuk YEARLY)
	// Karena Huma ga support pointer, kita validasi dengan cara lain:
	// User HARUS explicitly set year (ga bisa rely on default 0)
	if req.Year == 0 {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'year' wajib diisi (contoh: year=2025)")), nil
	}

	if req.Year < 2020 || req.Year > 2100 {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'year' harus antara 2020-2100")), nil
	}

	// Validate based on period type
	month := 0
	week := 0

	switch req.Period {
	case "DAILY":
		if req.Month == 0 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' wajib diisi untuk period DAILY (contoh: month=12)")), nil
		}
		if req.Month < 1 || req.Month > 12 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' harus antara 1-12")), nil
		}
		month = req.Month

		// Week is optional for DAILY
		if req.Week != 0 {
			if req.Week < 1 || req.Week > 5 {
				return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'week' harus antara 1-5")), nil
			}
			week = req.Week
		}

	case "WEEKLY":
		if req.Month == 0 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' wajib diisi untuk period WEEKLY (contoh: month=12)")), nil
		}
		if req.Month < 1 || req.Month > 12 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' harus antara 1-12")), nil
		}
		month = req.Month

	case "MONTHLY":
		// Year is enough for MONTHLY

	case "YEARLY":
		// Year is enough for YEARLY
	}

	period := model.ForecastPeriod(req.Period)

	data, err := h.svc.GetForecastsByPeriod(ctx, period, req.Year, month, week)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessForecastRetrieved), nil
}
