package user

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/user/model"
)

// UserService defines the interface for user-related operations.
type UserService interface {
	GetUserByEmail(ctx context.Context, email string) (*model.AuthUser, error)
	UpdateUserPassword(ctx context.Context, password string) error
}
