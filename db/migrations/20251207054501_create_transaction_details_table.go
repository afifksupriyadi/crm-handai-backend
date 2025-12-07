package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create transaction details table")

		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS transaction_details (
				id SERIAL PRIMARY KEY,
				transaction_code VARCHAR(20) NOT NULL REFERENCES transactions(code) ON DELETE CASCADE,
				product_id INTEGER NOT NULL NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
				variant_id INTEGER REFERENCES variants(id) ON DELETE RESTRICT,
				quantity INTEGER NOT NULL CHECK (quantity > 0),
				unit_price DECIMAL(10,2) NOT NULL,
				subtotal DECIMAL(10,2) NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ,
				deleted_at TIMESTAMPTZ
			);

			CREATE INDEX IF NOT EXISTS idx_transaction_details_transaction_code ON transaction_details(transaction_code);
			CREATE INDEX IF NOT EXISTS idx_transaction_details_product_id ON transaction_details(product_id);
			CREATE INDEX IF NOT EXISTS idx_transaction_details_variant_id ON transaction_details(variant_id);
			CREATE INDEX IF NOT EXISTS idx_transaction_details_deleted_at ON transaction_details(deleted_at);
		`)
		if err != nil {
			return fmt.Errorf("failed to create transaction details table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] drop transaction details table")

		_, err := db.ExecContext(ctx, `
			DROP TABLE IF EXISTS transaction_details CASCADE
		`)
		if err != nil {
			return fmt.Errorf("failed to drop transaction details table: %w", err)
		}

		return nil
	})
}
