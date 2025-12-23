package handler

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/request"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

type AnalyticsHandler struct {
	svc analytics.AnalyticsService
}

func NewAnalyticsHandler(svc analytics.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		svc: svc,
	}
}

// HandleGetSalesChart handles the sales chart request
func (h *AnalyticsHandler) HandleGetSalesChart(ctx context.Context, req *struct {
	request.AuthorizedRequest
	model.SalesChartRequest
}) (*response.Response, error) {
	// Validate request
	if err := req.SalesChartRequest.Validate(ctx); err != nil {
		return response.BuildError(ctx, err), nil
	}

	// Get sales chart data
	data, err := h.svc.GetSalesChart(ctx, &req.SalesChartRequest)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessSalesChart), nil
}

// HandleGetLoyalCustomers handles the loyal customers list request
func (h *AnalyticsHandler) HandleGetLoyalCustomers(ctx context.Context, req *model.GetLoyalCustomersRequest) (*response.Response, error) {
	// Get loyal customers
	data, err := h.svc.GetLoyalCustomers(ctx, req.Limit)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessCustomerRetrieved), nil
}

// HandleGetChurnCustomers handles the churn customers list request
func (h *AnalyticsHandler) HandleGetChurnCustomers(ctx context.Context, req *model.GetChurnCustomersRequest) (*response.Response, error) {
	// Get churn customers
	data, err := h.svc.GetChurnCustomers(ctx, req.Limit)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessCustomerRetrieved), nil
}
