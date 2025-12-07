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

	// User module
	userRepository "github.com/afifksupriyadi/crm-handai-backend/internal/modules/user/repository"
	userService "github.com/afifksupriyadi/crm-handai-backend/internal/modules/user/service"

	// Auth module
	authHandler "github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/handler"
	authRoutes "github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/routes"
	authService "github.com/afifksupriyadi/crm-handai-backend/internal/modules/auth/service"

	// Customer module
	customerRepository "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/repository"
	customerService "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/service"

	// Product module
	productRepository "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/repository"
	productService "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/service"

	// Transaction module
	transactionRepository "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/repository"
	transactionService "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/service"

	// Import module
	importHandler "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/handler"
	importRepository "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/repository"
	importRoutes "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/routes"
	importService "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/service"
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
	customerRepo := customerRepository.NewCustomerRepository(dbConn)
	productRepo := productRepository.NewProductRepository(dbConn)
	variantRepo := productRepository.NewVariantRepository(dbConn)
	transactionRepo := transactionRepository.NewTransactionRepository(dbConn)
	transactionDetailRepo := transactionRepository.NewTransactionDetailRepository(dbConn)
	importLogRepo := importRepository.NewImportLogRepository(dbConn)

	// Register services
	userSvc := userService.NewUserService(userRepo)
	customerSvc := customerService.NewCustomerService(customerRepo)
	productSvc := productService.NewProductService(productRepo)
	variantSvc := productService.NewVariantService(variantRepo)
	transactionSvc := transactionService.NewTransactionService(transactionRepo)
	transactionDetailSvc := transactionService.NewTransactionDetailService(transactionDetailRepo)
	importSvc := importService.NewImportService(customerSvc, productSvc, variantSvc, transactionSvc, transactionDetailSvc, importLogRepo)
	authSvc := authService.NewAuthService(c, userSvc)

	// Register handlers
	authHdlr := authHandler.NewAuthHandler(authSvc)
	importHdlr := importHandler.NewImportHandler(importSvc)

	// Register routes
	authRoutes.RegisterAuthRoutes(api, authHdlr)
	importRoutes.RegisterImportRoutes(f, importHdlr)

	// Register custom Huma error handler
	response.RegisterHumaErrorHandler()

	return api
}
