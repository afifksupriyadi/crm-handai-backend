package handler

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/request"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

// AuthHandler handles authentication-related API requests.
type AuthHandler struct {
	svc auth.AuthService
}

// NewAuthHandler initializes and returns a new AuthHandler instance.
func NewAuthHandler(svc auth.AuthService) *AuthHandler {
	return &AuthHandler{
		svc: svc,
	}
}

// HandleLogin processes user login requests.
func (h *AuthHandler) HandleLogin(ctx context.Context, req *request.GenericBodyRequest[model.LoginRequest]) (*response.Response, error) {
	body := req.Body
	if err := request.RequireBody(ctx, body); err != nil {
		return response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrEmptyRequestBody, "")), nil
	}

	body.Sanitize()
	if err := body.Validate(ctx); err != nil {
		return response.BuildError(ctx, err), nil
	}

	data, err := h.svc.Login(ctx, req.Body)
	if err != nil {
		return response.BuildError(ctx, err), nil
	}

	return response.BuildSuccess(data, response.SuccessLogin), nil
}
