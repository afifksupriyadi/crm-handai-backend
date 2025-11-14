package transport

import (
	"strings"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
	// app.Use()

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
	err, ok := e.(error)
	if !ok {
		err = errors.Errorf("%v", e)
	}

	requestID, _ := ctx.Locals("request_id").(string)

	log.Error().
		Str("request_id", requestID).
		Str("path", ctx.Path()).
		Str("method", ctx.Method()).
		Err(errors.WithStack(err)).
		Msg("Panic recovered")
}
