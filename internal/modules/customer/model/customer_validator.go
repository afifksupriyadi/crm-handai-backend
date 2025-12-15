package model

import (
	"context"
	"regexp"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

// Regular expressions for validation
var (
	phoneRegex = regexp.MustCompile(`^(\+62|62|0)[0-9]{9,12}$`)
)

// validateName validates that name is not empty and within length limit
func validateName(ctx context.Context, name string) error {
	if name == "" {
		return response.WrapAppError(ctx, nil, response.ErrEmptyName, "Name is required")
	}

	if len(name) > 50 {
		return response.WrapAppError(ctx, nil, response.ErrNameTooLong, "Name must not exceed 50 characters")
	}

	return nil
}

// validatePhone validates that phone is not empty and in correct format
func validatePhone(ctx context.Context, phone string) error {
	if phone == "" {
		return response.WrapAppError(ctx, nil, response.ErrEmptyPhone, "Phone is required")
	}

	if !phoneRegex.MatchString(phone) {
		return response.WrapAppError(ctx, nil, response.ErrInvalidPhoneFormat, "Phone format is invalid. Must be Indonesian phone number")
	}

	if len(phone) > 20 {
		return response.WrapAppError(ctx, nil, response.ErrPhoneTooLong, "Phone must not exceed 20 characters")
	}

	return nil
}
