package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create batches table")

		query := `
			CREATE SEQUENCE IF NOT EXISTS batches_id_seq;

			CREATE TABLE IF NOT EXISTS batches (
				id INTEGER NOT NULL DEFAULT nextval('batches_id_seq'::regclass),
				batch_date DATE NOT NULL,
				batch_code VARCHAR(50) NOT NULL UNIQUE,
				status VARCHAR(20) NOT NULL CHECK (status IN ('PROCESSING', 'COMPLETED', 'FAILED')),
				is_active BOOLEAN DEFAULT FALSE,
				customer_import_id INTEGER REFERENCES import_logs(id) ON DELETE SET NULL,
				transaction_import_id INTEGER REFERENCES import_logs(id) ON DELETE SET NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ,
				PRIMARY KEY (id)
			);

			CREATE INDEX IF NOT EXISTS idx_batches_date ON batches(batch_date);
			CREATE INDEX IF NOT EXISTS idx_batches_status ON batches(status);
			CREATE INDEX IF NOT EXISTS idx_batches_code ON batches(batch_code);
			CREATE UNIQUE INDEX IF NOT EXISTS idx_batches_single_active 
				ON batches (is_active) WHERE is_active = TRUE;

			ALTER SEQUENCE batches_id_seq OWNED BY batches.id;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create batches table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop batches table")

		query := `
			DROP INDEX IF EXISTS idx_batches_date;
			DROP INDEX IF EXISTS idx_batches_status;
			DROP INDEX IF EXISTS idx_batches_code;
			DROP INDEX IF EXISTS idx_batches_single_active;
			DROP TABLE IF EXISTS batches CASCADE;
			DROP SEQUENCE IF EXISTS batches_id_seq;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop batches table: %w", err)
		}

		return nil
	})
}
