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

	// Health module
	healthHandler "github.com/afifksupriyadi/crm-handai-backend/internal/modules/health/handler"
	healthRoutes "github.com/afifksupriyadi/crm-handai-backend/internal/modules/health/routes"

	// Customer module
	customerHandler "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/handler"
	customerRepository "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/repository"
	customerRoutes "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/routes"
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

	// Analytics module
	analyticsHandler "github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/handler"
	analyticsRepository "github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/repository"
	analyticsRoutes "github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/routes"
	analyticsService "github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/service"
)

func RegisterRoutes(f *fiber.App) huma.API {
	c := config.Get()

	// API configuration
	cfg := huma.DefaultConfig(c.ServiceName, "0.0.1")

	// Add JWT Bearer authentication scheme
	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Enter your JWT token in the format: Bearer <token>",
		},
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

	// Register Repositories
	userRepo := userRepository.NewUserRepository(dbConn)
	customerRepo := customerRepository.NewCustomerRepository(dbConn)
	productRepo := productRepository.NewProductRepository(dbConn)
	variantRepo := productRepository.NewVariantRepository(dbConn)
	transactionRepo := transactionRepository.NewTransactionRepository(dbConn)
	transactionDetailRepo := transactionRepository.NewTransactionDetailRepository(dbConn)
	customerBatchRepo := importRepository.NewCustomerBatchRepository(dbConn)
	transactionBatchRepo := importRepository.NewTransactionBatchRepository(dbConn)
	importLogRepo := importRepository.NewImportLogRepository(dbConn)
	importTrackerRepo := importRepository.NewImportTrackerRepository(dbConn)
	customerPredictionRepo := customerRepository.NewCustomerPredictionRepository(dbConn)
	customerSegmentRepo := customerRepository.NewCustomerSegmentRepository(dbConn)
	analyticsRepo := analyticsRepository.NewAnalyticsRepository(dbConn)
	customerPredictedProductRepo := customerRepository.NewCustomerPredictedProductRepository(dbConn)

	// Register Services
	windowCalculatorSvc := customerService.NewWindowCalculatorService()
	predictionCalculatorSvc := customerService.NewPredictionCalculatorService(customerPredictionRepo)
	predictionValidatorSvc := customerService.NewPredictionValidatorService(customerPredictionRepo)
	segmentDeterminerSvc := customerService.NewSegmentDeterminerService(customerPredictionRepo, customerSegmentRepo)
	productPredictionCalculatorSvc := customerService.NewProductPredictionCalculatorService(customerPredictedProductRepo)

	userSvc := userService.NewUserService(userRepo)
	customerSvc := customerService.NewCustomerService(customerRepo, dbConn)
	predictionOrchestratorSvc := customerService.NewPredictionOrchestratorService(
		dbConn,
		importTrackerRepo,
		customerPredictionRepo,
		customerSegmentRepo,
		customerPredictedProductRepo,
		customerSvc,
		windowCalculatorSvc,
		predictionCalculatorSvc,
		productPredictionCalculatorSvc,
		predictionValidatorSvc,
		segmentDeterminerSvc,
	)
	productSvc := productService.NewProductService(productRepo)
	variantSvc := productService.NewVariantService(variantRepo)
	transactionSvc := transactionService.NewTransactionService(transactionRepo)
	transactionDetailSvc := transactionService.NewTransactionDetailService(transactionDetailRepo)
	importSvc := importService.NewImportService(
		dbConn,
		customerSvc,
		productSvc,
		variantSvc,
		transactionSvc,
		transactionDetailSvc,
		customerBatchRepo,
		transactionBatchRepo,
		importLogRepo,
		importTrackerRepo,
		predictionOrchestratorSvc,
	)
	authSvc := authService.NewAuthService(c, userSvc)
	analyticsSvc := analyticsService.NewAnalyticsService(analyticsRepo)

	// Register Handlers
	authHdlr := authHandler.NewAuthHandler(authSvc)
	customerHdlr := customerHandler.NewCustomerHandler(customerSvc)
	importHdlr := importHandler.NewImportHandler(importSvc)
	healthHdlr := healthHandler.NewHealthHandler("0.0.1")
	analyticsHdlr := analyticsHandler.NewAnalyticsHandler(analyticsSvc)

	// Register Routes
	healthRoutes.RegisterHealthRoutes(api, healthHdlr)
	authRoutes.RegisterAuthRoutes(api, authHdlr)
	customerRoutes.RegisterCustomerRoutes(api, customerHdlr)
	importRoutes.RegisterImportRoutes(f, importHdlr)
	analyticsRoutes.RegisterAnalyticsRoutes(api, analyticsHdlr)

	// Register custom Huma error handler
	response.RegisterHumaErrorHandler()

	return api
}
