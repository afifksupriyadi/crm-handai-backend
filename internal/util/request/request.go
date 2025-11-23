package request

import (
	"context"
	"fmt"
)

// AuthorizedRequest is a base struct for requests that require authorization.
type AuthorizedRequest struct {
	Authorization string `header:"Authorization" validate:"required"`
}

// GenericRequest is a struct that can be used for requests that require an authorization header and an optional body.
type GenericRequest[T any] struct {
	AuthorizedRequest
	Body *T `json:"body,omitempty"`
}

// GenericRequestWithIDPath is a reusable request struct for endpoints that require an ID in the path and a request body.
type GenericRequestWithIDPath[T any] struct {
	AuthorizedRequest
	ID   int `path:"id" validate:"required"`
	Body *T  `json:"body,omitempty"`
}

// GenericBodyRequest is a reusable request struct for endpoints that require a request body.
type GenericBodyRequest[T any] struct {
	Body *T `json:"body,omitempty"`
}

// RequireBody returns an error if the request body is nil.
func RequireBody[T any](ctx context.Context, req *T) error {
	if req == nil {
		return fmt.Errorf("request body is required")
	}
	return nil
}
