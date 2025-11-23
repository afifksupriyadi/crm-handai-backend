package auth

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/model"
)

// AuthService defines the contract for authentication-related operations.
type AuthService interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
}
