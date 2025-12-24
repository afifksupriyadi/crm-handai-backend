package handler

import (
	"context"

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

// Request for customer detail with optional year and month filter
type GetCustomerDetailRequest struct {
	request.AuthorizedRequest
	ID    int `path:"id" validate:"required" doc:"Customer ID"`
	Year  int `query:"year" minimum:"0" doc:"Filter by year (e.g., 2025). If 0, all years"`
	Month int `query:"month" minimum:"0" maximum:"12" doc:"Filter by month (1-12). If 0, all months"`
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

// HandleGetCustomerByID now returns detailed information with optional year/month filter
func (h *CustomerHandler) HandleGetCustomerByID(ctx context.Context, req *GetCustomerDetailRequest) (*response.Response, error) {
	// Convert 0 values to nil pointers (0 = not provided)
	var yearPtr *int
	var monthPtr *int

	if req.Year != 0 {
		yearPtr = &req.Year
	}

	if req.Month != 0 {
		monthPtr = &req.Month
	}

	// Get detailed customer information with filters
	data, err := h.svc.GetCustomerDetail(ctx, req.ID, yearPtr, monthPtr)
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
