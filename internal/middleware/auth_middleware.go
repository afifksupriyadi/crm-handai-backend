package middleware

import (
	"context"
	"strings"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/security"
	"github.com/danielgtaylor/huma/v2"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserIDKey is the context key for storing user ID
	UserIDKey contextKey = "user_id"
	// UserNameKey is the context key for storing user name
	UserNameKey contextKey = "user_name"
	// UserVersionKey is the context key for storing token version
	UserVersionKey contextKey = "user_version"
)

// AuthMiddleware validates JWT token and injects user data into context
func AuthMiddleware(ctx huma.Context, next func(huma.Context)) {
	cfg := config.Get()

	// Get Authorization header
	authHeader := ctx.Header("Authorization")
	if authHeader == "" {
		response.WriteError(ctx, response.WrapAppError(
			ctx.Context(),
			nil,
			response.ErrTokenNotFound,
			"Authorization header is missing",
		))
		return
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		response.WriteError(ctx, response.WrapAppError(
			ctx.Context(),
			nil,
			response.ErrInvalidToken,
			"Invalid authorization header format",
		))
		return
	}

	tokenString := parts[1]

	// Validate token
	token, err := security.ValidateToken(
		tokenString,
		cfg.JWT.JWTSecretKey,
		cfg.JWT.JWTIssuer,
		cfg.JWT.JWTSignAlgorithm,
	)
	if err != nil {
		response.WriteError(ctx, response.WrapAppError(
			ctx.Context(),
			err,
			response.ErrInvalidToken,
			"Token validation failed",
		))
		return
	}

	// Extract claims
	claims := security.ExtractClaims(token)

	// Inject user data into context
	newCtx := context.WithValue(ctx.Context(), UserIDKey, claims.UserID)
	newCtx = context.WithValue(newCtx, UserNameKey, claims.Name)
	newCtx = context.WithValue(newCtx, UserVersionKey, claims.Version)

	// Update context
	ctx = huma.WithContext(ctx, newCtx)

	next(ctx)
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

// GetUserNameFromContext retrieves user name from context
func GetUserNameFromContext(ctx context.Context) (string, bool) {
	userName, ok := ctx.Value(UserNameKey).(string)
	return userName, ok
}

// GetUserVersionFromContext retrieves token version from context
func GetUserVersionFromContext(ctx context.Context) (int, bool) {
	version, ok := ctx.Value(UserVersionKey).(int)
	return version, ok
}
