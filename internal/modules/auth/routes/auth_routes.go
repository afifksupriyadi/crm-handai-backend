package routes

import (
	"fmt"
	"net/http"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/middleware"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/handler"
	"github.com/danielgtaylor/huma/v2"
)

// RegisterAuthRoutes registers the authentication routes with the provided API and handler.
func RegisterAuthRoutes(api huma.API, h *handler.AuthHandler) {
	basePath := fmt.Sprintf("%s/auth", config.Get().BasePath)

	// POST /auth/login - Public endpoint (no auth required)
	huma.Register(api,
		huma.Operation{
			OperationID: "login",
			Method:      http.MethodPost,
			Path:        basePath + "/login",
			Summary:     "Login",
			Description: "Login a user and returns a JWT access token.",
			Tags:        []string{"auth"},
			Middlewares: huma.Middlewares{},
		}, h.HandleLogin,
	)

	// GET /auth/me - Protected endpoint (requires JWT)
	huma.Register(api,
		huma.Operation{
			OperationID: "getCurrentUser",
			Method:      http.MethodGet,
			Path:        basePath + "/me",
			Summary:     "Get Current User",
			Description: "Retrieves the authenticated user's profile information.",
			Tags:        []string{"auth"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Middlewares: huma.Middlewares{
				middleware.AuthMiddleware,
			},
		}, h.HandleGetCurrentUser,
	)
}
