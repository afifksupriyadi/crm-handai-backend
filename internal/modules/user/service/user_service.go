package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/user"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/user/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/user/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

// UserServiceImpl implements the UserService interface.
type UserServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new instance of UserServiceImpl.
func NewUserService(userRepo repository.UserRepository) user.UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

// GetUserByEmail retrieves user by email from repository.
func (s *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*model.AuthUser, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrUserNotFound, "User not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get user from database")
	}

	return user, nil
}

// GetUserByID retrieves user by ID from repository.
func (s *UserServiceImpl) GetUserByID(ctx context.Context, userID int) (*model.AuthUser, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrUserNotFound, "User not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get user from database")
	}

	return user, nil
}

func (s *UserServiceImpl) UpdateUserPassword(ctx context.Context, password string) error {
	// TODO: implement update password service
	return nil
}
