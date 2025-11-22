package migration

import (
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/db/migrations"
	"github.com/afifksupriyadi/crm-handai-backend/lib/db"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun/migrate"
)

// InitCmd is command to create migration tables.
var InitCmd = &cobra.Command{
	Use:   "db:init",
	Short: "Create migration tables",
	Long:  `Initialize database migration tables (bun_migrations and bun_migration_locks).`,
}

// dbInit initializes the database migration tables.
func dbInit(cmd *cobra.Command, args []string) error {
	c := config.Get()

	dbConn, err := db.Open(c)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer dbConn.Close()

	migrator := migrate.NewMigrator(dbConn, migrations.Migrations)

	if err := migrator.Init(cmd.Context()); err != nil {
		return fmt.Errorf("failed to initialize migration tables: %w", err)
	}

	fmt.Println("Migration tables created successfully")
	return nil
}

func init() {
	InitCmd.RunE = dbInit
}
