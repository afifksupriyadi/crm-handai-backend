package model

import (
	"context"
	"regexp"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

// Regular expressions for validation
var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// IsValidEmail checks if the provided string is a valid email address.
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// validateEmail validates that email is not empty and is in correct format.
func validateEmail(ctx context.Context, email string) error {
	if email == "" {
		return response.WrapAppError(ctx, nil, response.ErrEmptyEmail, "Email is not provided")
	}

	if !IsValidEmail(email) {
		return response.WrapAppError(ctx, nil, response.ErrInvalidEmailFormat, "Invalid email format")
	}

	return nil
}

// validatePassword validates that password is not empty.
func validatePassword(ctx context.Context, password string) error {
	if password == "" {
		return response.WrapAppError(ctx, nil, response.ErrEmptyPassword, "Password is not provided")
	}

	return nil
}
