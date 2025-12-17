package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] creating analytics schema...")

		// Create analytics schema
		_, err := db.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS analytics`)
		if err != nil {
			return fmt.Errorf("failed to create analytics schema: %w", err)
		}

		fmt.Println(" done")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] dropping analytics schema...")

		// Drop analytics schema (CASCADE will drop all tables in it)
		_, err := db.ExecContext(ctx, `DROP SCHEMA IF EXISTS analytics CASCADE`)
		if err != nil {
			return fmt.Errorf("failed to drop analytics schema: %w", err)
		}

		fmt.Println(" done")
		return nil
	})
}
