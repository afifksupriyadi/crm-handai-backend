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
	Period string `query:"period" doc:"Forecast period (DAILY, WEEKLY, MONTHLY, YEARLY)"`
	Year   *int   `query:"year" doc:"Year (2020-2100) - Required for all periods"`
	Month  *int   `query:"month" doc:"Month (1-12) - Required for DAILY, WEEKLY, MONTHLY"`
	Week   *int   `query:"week" doc:"Week number (1-5) - Optional for DAILY"`
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
	if req.Year == nil {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'year' wajib diisi untuk semua period")), nil
	}

	year := *req.Year
	if year < 2020 || year > 2100 {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'year' harus antara 2020-2100")), nil
	}

	// Validate based on period type
	month := 0
	week := 0

	switch req.Period {
	case "DAILY":
		if req.Month == nil {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' wajib diisi untuk period DAILY")), nil
		}
		month = *req.Month
		if month < 1 || month > 12 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' harus antara 1-12")), nil
		}
		// Week is optional for DAILY
		if req.Week != nil {
			week = *req.Week
			if week < 1 || week > 5 {
				return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'week' harus antara 1-5")), nil
			}
		}

	case "WEEKLY":
		if req.Month == nil {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' wajib diisi untuk period WEEKLY")), nil
		}
		month = *req.Month
		if month < 1 || month > 12 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' harus antara 1-12")), nil
		}

	case "MONTHLY":
		// Year is enough for MONTHLY

	case "YEARLY":
		// Year is enough for YEARLY
	}

	period := model.ForecastPeriod(req.Period)

	data, err := h.svc.GetForecastsByPeriod(ctx, period, year, month, week)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessForecastRetrieved), nil
}
