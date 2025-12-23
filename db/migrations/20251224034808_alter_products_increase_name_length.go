package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] increase products.name column length from VARCHAR(30) to VARCHAR(100)")

		_, err := db.ExecContext(ctx, `
			ALTER TABLE products
			ALTER COLUMN name TYPE VARCHAR(100);
		`)
		if err != nil {
			return fmt.Errorf("failed to alter products.name column: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] revert products.name column length to VARCHAR(30)")

		_, err := db.ExecContext(ctx, `
			ALTER TABLE products
			ALTER COLUMN name TYPE VARCHAR(30);
		`)
		if err != nil {
			return fmt.Errorf("failed to revert products.name column: %w", err)
		}

		return nil
	})
}
