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
	Period string `query:"period" doc:"Forecast period (WEEKLY, MONTHLY, YEARLY)"`
	Year   int    `query:"year" doc:"Year (2020-2100)"`
	Month  int    `query:"month" default:"0" doc:"Month (1-12) - Required for WEEKLY & MONTHLY"`
	Week   int    `query:"week" default:"0" doc:"Week number in month (1-5) - Required for WEEKLY only"`
}

func (h *ForecastingHandler) HandleGetForecasts(ctx context.Context, req *GetSalesForecastQueryRequest) (*response.Response, error) {
	// Validate period
	if req.Period == "" {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'period' wajib diisi")), nil
	}

	validPeriods := map[string]bool{"WEEKLY": true, "MONTHLY": true, "YEARLY": true}
	if !validPeriods[req.Period] {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'period' tidak valid (harus WEEKLY, MONTHLY, atau YEARLY)")), nil
	}

	// Validate year
	if req.Year == 0 {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'year' wajib diisi")), nil
	}

	if req.Year < 2020 || req.Year > 2100 {
		return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'year' harus antara 2020-2100")), nil
	}

	// Validate based on period type
	month := 0
	week := 0

	switch req.Period {
	case "WEEKLY":
		if req.Month == 0 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' wajib diisi untuk period WEEKLY")), nil
		}
		if req.Week == 0 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'week' wajib diisi untuk period WEEKLY")), nil
		}
		if req.Month < 1 || req.Month > 12 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' harus antara 1-12")), nil
		}
		if req.Week < 1 || req.Week > 5 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'week' harus antara 1-5")), nil
		}
		month = req.Month
		week = req.Week

	case "MONTHLY":
		if req.Month == 0 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' wajib diisi untuk period MONTHLY")), nil
		}
		if req.Month < 1 || req.Month > 12 {
			return response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrUnprocessableEntity, "Parameter 'month' harus antara 1-12")), nil
		}
		month = req.Month

	case "YEARLY":
		// No additional validation needed
	}

	period := model.ForecastPeriod(req.Period)

	data, err := h.svc.GetForecastsByPeriod(ctx, period, req.Year, month, week)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessForecastRetrieved), nil
}
