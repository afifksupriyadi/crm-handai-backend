package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create transactions table")

		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS transactions (
				code VARCHAR(20) PRIMARY KEY,
				customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
				transaction_date TIMESTAMPTZ NOT NULL,
				discount DECIMAL(10,2) NOT NULL DEFAULT 0,
				shipping_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
				payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('Tunai', 'Non Tunai')),
				status VARCHAR(20) NOT NULL CHECK (status IN ('LUNAS', 'PENDING')),
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ,
				deleted_at TIMESTAMPTZ
			);

			CREATE INDEX IF NOT EXISTS idx_transactions_customer_id ON transactions(customer_id);
			CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date);
			CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
			CREATE INDEX IF NOT EXISTS idx_transactions_deleted_at ON transactions(deleted_at);
		`)
		if err != nil {
			return fmt.Errorf("failed to create transactions table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop transactions table")

		_, err := db.ExecContext(ctx, `
			DROP TABLE IF EXISTS transactions CASCADE
		`)
		if err != nil {
			return fmt.Errorf("failed to drop transactions table: %w", err)
		}

		return nil
	})
}
