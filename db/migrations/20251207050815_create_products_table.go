package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create products table")

		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS products (
				id SERIAL PRIMARY KEY,
				name VARCHAR(30) NOT NULL,
				category VARCHAR(20) NOT NULL CHECK (category IN ('COFFEE', 'NON-COFFEE')),
				base_price DECIMAL(10,2) NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ,
				deleted_at TIMESTAMPTZ
			);

			CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
			CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
			CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at);
		`)
		if err != nil {
			return fmt.Errorf("failed to create products table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop products table")

		_, err := db.ExecContext(ctx, `
			DROP TABLE IF EXISTS products CASCADE
		`)
		if err != nil {
			return fmt.Errorf("failed to drop products table: %w", err)
		}

		return nil
	})
}
