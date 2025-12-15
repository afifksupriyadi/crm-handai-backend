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

// HandleGetCustomerByID processes get customer by ID requests
func (h *CustomerHandler) HandleGetCustomerByID(ctx context.Context, req *request.GenericRequestWithIDPath[any]) (*response.Response, error) {
	data, err := h.svc.GetCustomerByID(ctx, req.ID)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessCustomerRetrieved), nil
}

// HandleGetAllCustomers processes get all customers requests
func (h *CustomerHandler) HandleGetAllCustomers(ctx context.Context, req *request.GenericRequest[model.GetCustomersRequest]) (*response.Response, error) {
	// Use query params directly since it's GET request
	queryReq := &model.GetCustomersRequest{
		Page:   1,
		Limit:  10,
		Search: "",
	}

	// If Body is not nil, use its values
	if req.Body != nil {
		queryReq = req.Body
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
