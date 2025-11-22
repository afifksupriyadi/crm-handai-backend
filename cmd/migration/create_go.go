package migration

import (
	"fmt"
	"strings"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/db/migrations"
	"github.com/afifksupriyadi/crm-handai-backend/lib/db"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun/migrate"
)

// CreateGoCmd is command to create Go migration.
var CreateGoCmd = &cobra.Command{
	Use:   "db:create_go [migration_name]",
	Short: "Create Go migration",
	Long: `Create a new Go migration file with the specified name.
Example:
  db:create_go create_users_table
  db:create_go add_customers_table`,
	Args: cobra.MinimumNArgs(1),
}

// createGo creates a new Go migration file with the specified name.
func createGo(cmd *cobra.Command, args []string) error {
	c := config.Get()

	dbConn, err := db.Open(c)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer dbConn.Close()

	migrator := migrate.NewMigrator(dbConn, migrations.Migrations)

	name := strings.Join(args, "_")
	mf, err := migrator.CreateGoMigration(cmd.Context(), name)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	fmt.Printf("Created migration: %s\n", mf.Name)
	fmt.Printf("Path: %s\n", mf.Path)
	return nil
}

func init() {
	CreateGoCmd.RunE = createGo
}
