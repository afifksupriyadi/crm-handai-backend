package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create variants table")

		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS variants (
				id SERIAL PRIMARY KEY,
				product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
				name VARCHAR(30) NOT NULL,
				price_modifier DECIMAL(10,2) NOT NULL DEFAULT 0,
				is_default BOOLEAN NOT NULL DEFAULT FALSE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ,
				deleted_at TIMESTAMPTZ
			);

			CREATE INDEX IF NOT EXISTS idx_variants_product_id 
				ON variants(product_id);

			CREATE INDEX IF NOT EXISTS idx_variants_deleted_at 
				ON variants(deleted_at);
			
			CREATE UNIQUE INDEX IF NOT EXISTS idx_variants_product_name 
				ON variants(product_id, name) WHERE deleted_at IS NULL;
		`)
		if err != nil {
			return fmt.Errorf("failed to create variants table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop variants table")

		_, err := db.ExecContext(ctx, `
			DROP TABLE IF EXISTS variants CASCADE
		`)
		if err != nil {
			return fmt.Errorf("failed to drop variants table: %w", err)
		}

		return nil
	})
}
