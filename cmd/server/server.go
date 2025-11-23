package server

import (
	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/middleware"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/afifksupriyadi/crm-handai-backend/lib/db"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"

	userRepository "github.com/afifksupriyadi/crm-handai-backend/internal/modules/user/repository"
	userService "github.com/afifksupriyadi/crm-handai-backend/internal/modules/user/service"

	authHandler "github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/handler"
	authRoutes "github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/routes"
	authService "github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/service"
)

func RegisterRoutes(f *fiber.App) huma.API {
	c := config.Get()

	// API configuration
	cfg := huma.DefaultConfig(c.ServiceName, "0.0.1")

	// Disable docs in production
	if c.IsProduction() {
		cfg.DocsPath = ""
		f.Get("/openapi.yaml", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusNotFound)
		})
	}

	cfg.Servers = []*huma.Server{
		{URL: c.PublishURL},
	}

	// create Huma API with fiber adapter
	api := humafiber.New(f, cfg)
	api.UseMiddleware(middleware.WrapFiberContextMiddleware)

	// database connection
	dbConn, err := db.Open(c)
	if err != nil {
		logger.Get().Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Register repositories
	userRepo := userRepository.NewUserRepository(dbConn)

	// Register services
	userSvc := userService.NewUserService(userRepo)
	authSvc := authService.NewAuthService(c, userSvc)

	// Register handlers
	authHdlr := authHandler.NewAuthHandler(authSvc)

	// Register routes
	authRoutes.RegisterAuthRoutes(api, authHdlr)

	// Register custom Huma error handler
	response.RegisterHumaErrorHandler()

	return api
}
