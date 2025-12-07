package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create customer table")

		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS customers (
				id SERIAL PRIMARY KEY,
				name VARCHAR(50) NOT NULL,
				phone VARCHAR(20) NOT NULL UNIQUE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ,
				deleted_at TIMESTAMPTZ
			);

			CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
			CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(name);
			CREATE INDEX IF NOT EXISTS idx_customers_deleted_at ON customers(deleted_at);
		`)
		if err != nil {
			return fmt.Errorf("failed to create customer table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop customers table")

		_, err := db.ExecContext(ctx, `
			DROP TABLE IF EXISTS customers CASCADE;
		`)
		if err != nil {
			return fmt.Errorf("failed to drop customers table: %w", err)
		}

		return nil
	})
}
