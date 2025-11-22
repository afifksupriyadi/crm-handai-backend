package migration

import (
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/db/migrations"
	"github.com/afifksupriyadi/crm-handai-backend/lib/db"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun/migrate"
)

// RollbackCmd is command to rollback the last migration group.
var RollbackCmd = &cobra.Command{
	Use:   "db:rollback",
	Short: "Rollback the last migration group",
	Long:  `Rollback the last applied migration group from the database.`,
}

// rollback handles the database migration rollback operation.
func rollback(cmd *cobra.Command, args []string) error {
	c := config.Get()

	dbConn, err := db.Open(c)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer dbConn.Close()

	migrator := migrate.NewMigrator(dbConn, migrations.Migrations)

	group, err := migrator.Rollback(cmd.Context())
	if err != nil {
		return fmt.Errorf("migration rollback error: %w", err)
	}

	if group.ID == 0 {
		fmt.Println("No migration groups available to roll back")
		return nil
	}

	fmt.Printf("Successfully rolled back migration group: %s\n", group)
	return nil
}

func init() {
	RollbackCmd.RunE = rollback
}
