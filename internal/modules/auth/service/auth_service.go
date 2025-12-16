package service

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/user"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/security"
)

// AuthServiceImpl provides the implementation of AuthService.
type AuthServiceImpl struct {
	conf    *config.Config
	userSvc user.UserService
}

// NewAuthService constructs a new AuthServiceImpl.
func NewAuthService(conf *config.Config, userSvc user.UserService) auth.AuthService {
	return &AuthServiceImpl{
		conf:    conf,
		userSvc: userSvc,
	}
}

// Login authenticates a user using their email and password, and returns a signed JWT token if successful.
func (s *AuthServiceImpl) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userSvc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if !security.ComparePassword(req.Password, user.PasswordHash) {
		return nil, response.WrapAppError(ctx, nil, response.ErrInvalidCredentials, "Invalid password")
	}

	jwtCfg := s.conf.JWT
	tokenParams := security.GenerateTokenParam{
		UserID:     user.ID,
		Name:       user.Name,
		Expiration: jwtCfg.JWTExpiration,
		Version:    jwtCfg.JWTVersion,
		Issuer:     jwtCfg.JWTIssuer,
		Algorithm:  jwtCfg.JWTSignAlgorithm,
		SecretKey:  jwtCfg.JWTSecretKey,
	}

	token, err := security.GenerateUserToken(tokenParams)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInternalServerError, "Failed to generate token")
	}

	logger.Get().Info().Msg("Login process successful")
	return &model.LoginResponse{Token: token}, nil
}

// GetCurrentUser retrieves the authenticated user's profile from the database.
func (s *AuthServiceImpl) GetCurrentUser(ctx context.Context, userID int) (*model.CurrentUserResponse, error) {
	user, err := s.userSvc.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &model.CurrentUserResponse{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Version: s.conf.JWT.JWTVersion,
	}, nil
}
