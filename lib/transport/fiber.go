package transport

import (
	"strings"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// InitFiber initializes and configures a new Fiber application instance.
// The function takes a config parameter to customize the application behavior
// and returns a configured *fiber.App instance ready to be used.
//
// Parameters:
//   - c: *config.Config - Application configuration containing service name settings
//
// Returns:
//   - *fiber.App - Configured Fiber application instance
func InitFiber(c *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               c.ServiceName,
	})

	// requestID middleware
	app.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.New().String()
		},
		ContextKey: "request_id",
	}))

	// logging middleware
	app.Use(loggingMiddleware(c))

	// recovery middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace:  true,
		StackTraceHandler: stackTreeHandler,
	}))

	// security header middleware

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// skip for huma docs
		if !strings.HasPrefix(c.Path(), "/docs") {
			c.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		}

		return c.Next()
	})

	// CORS middleware with improved configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins: c.CORS.AllowedOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
		MaxAge:       300,
	}))

	return app
}

func stackTreeHandler(ctx *fiber.Ctx, e interface{}) {
	log.Ctx(ctx.UserContext()).Error().
		Str("path", ctx.Path()).
		Str("method", ctx.Method()).
		Interface("panic", e).
		Msg("Panic recovered")
}

func loggingMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !cfg.Logger.AccessLog {
			return c.Next()
		}

		start := time.Now()
		requestID := c.Locals("request_id").(string)

		// add request_id to context
		ctx := logger.WithRequestID(c.UserContext(), requestID)
		c.SetUserContext(ctx)

		// log incoming request
		logEvent := log.Ctx(ctx).Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP())

		// log headers
		if cfg.Logger.LogHeaders {
			logEvent = logEvent.Interface("headers", c.GetReqHeaders())
		}

		// log query params
		if cfg.Logger.LogQueryParams {
			logEvent = logEvent.Interface("query", c.Queries())
		}

		logEvent.Msg("Incoming request")

		// process request
		err := c.Next()

		// log response
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()

		logEvent = log.Ctx(ctx).Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", statusCode).
			Dur("duration", duration).
			Int64("duration_ms", duration.Milliseconds())

		if cfg.Logger.LogBody && statusCode >= 400 {
			logEvent = logEvent.Str("req_body", string(c.Body()))
		}

		logEvent.Msg("Request completed")

		return err
	}
}
