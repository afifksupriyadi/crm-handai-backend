package logger

import (
	"os"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global logger with config
func Init(cfg *config.Config) {
	// set log level
	level := zerolog.InfoLevel
	if cfg.IsDevelopment() {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	// set output format
	if cfg.IsDevelopment() {
		// pretty console output for development
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		// JSON output for production
		log.Logger = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Caller().
			Str("service", cfg.ServiceName).
			Str("env", cfg.Env).
			Logger()
	}

	log.Info().Msg("Logger initialized")
}

// Get returns the global logger
func Get() *zerolog.Logger {
	return &log.Logger
}
