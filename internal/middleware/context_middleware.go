package middleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
)

// WrapFiberContextMiddleware ensures the Fiber context is properly propagated to Huma handlers.
// This middleware bridges Fiber's context with Huma's context system.
func WrapFiberContextMiddleware(ctx huma.Context, next func(huma.Context)) {
	// Get Fiber context from Huma's underlying context
	fiberCtx := ctx.Context().Value("fiber")

	if fiberCtx != nil {
		if fc, ok := fiberCtx.(*fiber.Ctx); ok {
			// Use UserContext from Fiber (which has our logger injected)
			ctx = huma.WithContext(ctx, fc.UserContext())
		}
	}

	next(ctx)
}
