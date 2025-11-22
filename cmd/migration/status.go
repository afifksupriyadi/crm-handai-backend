package migration

import (
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/db/migrations"
	"github.com/afifksupriyadi/crm-handai-backend/lib/db"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun/migrate"
)

// StatusCmd command to print migrations status.
var StatusCmd = &cobra.Command{
	Use:   "db:status",
	Short: "Print migrations status",
	Long:  `Print the status of all database migrations, showing which ones have been applied and which are pending.`,
}

// status prints the current status of database migrations.
func status(cmd *cobra.Command, args []string) error {
	c := config.Get()

	dbConn, err := db.Open(c)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer dbConn.Close()

	migrator := migrate.NewMigrator(dbConn, migrations.Migrations)

	ms, err := migrator.MigrationsWithStatus(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Printf("Migrations: %d total\n", len(ms))
	fmt.Println("---")

	for i, m := range ms {
		status := "pending"
		if m.GroupID > 0 {
			status = "applied"
		}
		fmt.Printf("%d. %s [%s]\n", i+1, m.Name, status)
	}

	fmt.Println("---")
	fmt.Printf("Unapplied migrations: %d\n", len(ms.Unapplied()))
	fmt.Printf("Last migration group: %s\n", ms.LastGroup())

	return nil
}

func init() {
	StatusCmd.RunE = status
}
