package migrations

import (
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun/migrate"
)

// Migrations is an object for initializing migrations.
var Migrations = migrate.NewMigrations()

func init() {
	if err := Migrations.DiscoverCaller(); err != nil {
		log.Error().Err(err).Msg("Failed to discover migrations")
	}
}
