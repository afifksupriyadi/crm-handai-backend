package handler

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/request"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

// CustomerHandler handles customer-related API requests
type CustomerHandler struct {
	svc customer.CustomerService
}

// NewCustomerHandler creates a new instance of CustomerHandler
func NewCustomerHandler(svc customer.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

// GetCustomersQueryRequest untuk GET request dengan query params
type GetCustomersQueryRequest struct {
	request.AuthorizedRequest
	Page      int    `query:"page" default:"1" minimum:"1" doc:"Page number for pagination"`
	Limit     int    `query:"limit" default:"10" minimum:"1" maximum:"100" doc:"Number of items per page"`
	Search    string `query:"search" default:"" doc:"Search by name or phone"`
	SortOrder string `query:"sort_order" default:"asc" enum:"asc,desc" doc:"Sort order by ID (asc/desc)"`
}

// Request for customer detail with optional month filter
type GetCustomerDetailRequest struct {
	request.AuthorizedRequest
	ID    int    `path:"id" validate:"required" doc:"Customer ID"`
	Month string `query:"month" doc:"Filter month (format: YYYY-MM, e.g., 2025-11)"`
}

// HandleCreateCustomer processes create customer requests
func (h *CustomerHandler) HandleCreateCustomer(ctx context.Context, req *request.GenericRequest[model.CreateCustomerRequest]) (*response.Response, error) {
	body := req.Body
	if err := request.RequireBody(ctx, body); err != nil {
		return response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrEmptyRequestBody, "")), nil
	}

	body.Sanitize()
	if err := body.Validate(ctx); err != nil {
		return response.BuildError(ctx, err), nil
	}

	data, err := h.svc.CreateCustomer(ctx, body)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessCustomerCreated), nil
}

// HandleGetCustomerByID now returns detailed information with optional month filter
func (h *CustomerHandler) HandleGetCustomerByID(ctx context.Context, req *GetCustomerDetailRequest) (*response.Response, error) {
	// Parse month filter if provided
	var monthFilter *time.Time
	if req.Month != "" {
		parsed, err := time.Parse("2006-01", req.Month)
		if err != nil {
			return response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidDateFormat, "Invalid month format. Use YYYY-MM")), nil
		}
		monthFilter = &parsed
	}

	// Get detailed customer information
	data, err := h.svc.GetCustomerDetail(ctx, req.ID, monthFilter)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessCustomerRetrieved), nil
}

// HandleGetAllCustomers processes get all customers requests
func (h *CustomerHandler) HandleGetAllCustomers(ctx context.Context, req *GetCustomersQueryRequest) (*response.Response, error) {
	queryReq := &model.GetCustomersRequest{
		Page:      req.Page,
		Limit:     req.Limit,
		Search:    req.Search,
		SortOrder: req.SortOrder,
	}

	data, err := h.svc.GetAllCustomers(ctx, queryReq)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessCustomerRetrieved), nil
}

// HandleUpdateCustomer processes update customer requests
func (h *CustomerHandler) HandleUpdateCustomer(ctx context.Context, req *request.GenericRequestWithIDPath[model.UpdateCustomerRequest]) (*response.Response, error) {
	body := req.Body
	if err := request.RequireBody(ctx, body); err != nil {
		return response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrEmptyRequestBody, "")), nil
	}

	body.Sanitize()
	if err := body.Validate(ctx); err != nil {
		return response.BuildError(ctx, err), nil
	}

	data, err := h.svc.UpdateCustomer(ctx, req.ID, body)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessCustomerUpdated), nil
}

// HandleDeleteCustomer processes delete customer requests
func (h *CustomerHandler) HandleDeleteCustomer(ctx context.Context, req *request.GenericRequestWithIDPath[any]) (*response.Response, error) {
	err := h.svc.DeleteCustomer(ctx, req.ID)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(struct{}{}, response.SuccessCustomerDeleted), nil
}

// HandleGetCustomersWithRecentTransactions processes get customers with recent transactions
func (h *CustomerHandler) HandleGetCustomersWithRecentTransactions(ctx context.Context, req *model.GetRecentTransactionsQueryRequest) (*response.Response, error) {
	serviceReq := &model.GetRecentTransactionsRequest{
		Page:  req.Page,
		Limit: req.Limit,
	}

	data, err := h.svc.GetCustomersWithRecentTransactions(ctx, serviceReq)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessCustomerRetrieved), nil
}

// HandleGetCustomerPredictions processes get customer predictions requests
func (h *CustomerHandler) HandleGetCustomerPredictions(ctx context.Context, req *model.GetCustomerPredictionsRequest) (*response.Response, error) {
	data, err := h.svc.GetCustomerPredictions(ctx, req.ID, req.Limit)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessCustomerRetrieved), nil
}
