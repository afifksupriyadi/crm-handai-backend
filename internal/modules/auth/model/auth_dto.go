package model

import (
	"context"
	"strings"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Sanitize removes leading/trailing spaces from the request fields.
func (r *LoginRequest) Sanitize() {
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)
}

// Validate runs all field-level validations for login request.
func (r *LoginRequest) Validate(ctx context.Context) error {
	if err := validateEmail(ctx, r.Email); err != nil {
		return err
	}

	if err := validatePassword(ctx, r.Password); err != nil {
		return err
	}

	return nil
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	Token string `json:"token"`
}
