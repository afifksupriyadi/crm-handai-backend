package cmd

import (
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/cmd/migration"
	"github.com/afifksupriyadi/crm-handai-backend/cmd/server"
	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/lib/transport"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

// Options represents the command line options for the application.
type Options struct {
	Host string `doc:"Hostname to listen on."`
	Port int    `doc:"Port to listen on." short:"p"`
}

// applyOptions applies command line options to the configuration.
func applyOptions(opts *Options) *config.Config {
	c := config.Get()

	if opts.Host != "" {
		c.Host = opts.Host
	}

	if opts.Port != 0 {
		c.Port = opts.Port
	}

	return c
}

// startServer launches the Fiber server.
func startServer(app *fiber.App, cfg *config.Config) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger.Get().Info().Str("address", addr).Msg("Starting server")

	if err := app.Listen(addr); err != nil {
		logger.Get().Fatal().Err(err).Msg("Failed to start server")
	}
}

// gracefulShutdown stops the Fiber app with timeout.
func gracefulShutdown(app *fiber.App, cfg *config.Config) {
	logger.Get().Info().Msg("Initiating graceful shutdown...")

	if err := app.ShutdownWithTimeout(cfg.ShutdownTimeout); err != nil {
		logger.Get().Error().Err(err).Msg("Error during graceful shutdown")
	}

	logger.Get().Info().Msg("Server shutdown completed")
}

// registerMigrationCommands adds all migration subcommands.
func registerMigrationCommands(root *cobra.Command) {
	root.AddCommand(
		migration.InitCmd,
		migration.MigrateCmd,
		migration.RollbackCmd,
		migration.StatusCmd,
		migration.CreateGoCmd,
	)
}

// Execute runs the main CLI entrypoint for the application.
func Execute() {
	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		cfg := applyOptions(opts)
		logger.Init(cfg)

		fiberApp := transport.InitFiber(cfg)

		hooks.OnStart(func() {
			server.RegisterRoutes(fiberApp)
			startServer(fiberApp, cfg)
		})

		hooks.OnStop(func() {
			gracefulShutdown(fiberApp, cfg)
		})
	})

	root := cli.Root()
	root.Use = config.Get().ServiceName
	root.Version = version

	registerMigrationCommands(root)

	cli.Run()
}
