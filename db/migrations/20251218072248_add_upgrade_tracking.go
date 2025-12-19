package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] add upgrade tracking to customers")

		query := `
			ALTER TABLE customers 
			ADD COLUMN IF NOT EXISTS upgraded_from_guest BOOLEAN NOT NULL DEFAULT false,
			ADD COLUMN IF NOT EXISTS upgraded_at TIMESTAMPTZ,
			ADD COLUMN IF NOT EXISTS first_seen_as_guest TIMESTAMPTZ;

			CREATE INDEX IF NOT EXISTS idx_customers_upgraded 
			ON customers(upgraded_from_guest) 
			WHERE upgraded_from_guest = true;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to add upgrade tracking to customers: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] remove upgrade tracking from customers")

		query := `
			DROP INDEX IF EXISTS idx_customers_upgraded;
			ALTER TABLE customers 
			DROP COLUMN IF EXISTS upgraded_from_guest,
			DROP COLUMN IF EXISTS upgraded_at,
			DROP COLUMN IF EXISTS first_seen_as_guest;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to remove upgrade tracking from customers: %w", err)
		}

		return nil
	})
}
