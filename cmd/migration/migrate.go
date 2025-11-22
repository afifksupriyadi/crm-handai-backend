package migration

import (
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/db/migrations"
	"github.com/afifksupriyadi/crm-handai-backend/lib/db"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun/migrate"
)

// MigrateCmd is command to migrate database.
var MigrateCmd = &cobra.Command{
	Use:   "db:migrate",
	Short: "Migrate database",
	Long: `Migrate database to the latest version.
This command will apply all pending migrations to the database.`,
}

// dbMigrate runs all pending migrations.
func dbMigrate(cmd *cobra.Command, args []string) error {
	c := config.Get()

	dbConn, err := db.Open(c)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer dbConn.Close()

	migrator := migrate.NewMigrator(dbConn, migrations.Migrations)

	group, err := migrator.Migrate(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if group.ID == 0 {
		fmt.Println("There are no new migrations to run")
		return nil
	}

	fmt.Printf("Successfully migrated to %s\n", group)
	return nil
}

func init() {
	MigrateCmd.RunE = dbMigrate
}
