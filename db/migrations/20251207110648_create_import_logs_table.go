package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] create import_logs table")

		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS import_logs (
				id SERIAL PRIMARY KEY,
				import_type VARCHAR(50) NOT NULL,
				file_date DATE NOT NULL,
				filename VARCHAR(255) NOT NULL,
				rows_imported INTEGER NOT NULL DEFAULT 0,
				status VARCHAR(20) NOT NULL,
				imported_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
			);

			CREATE INDEX IF NOT EXISTS idx_import_logs_type_date ON import_logs(import_type, file_date);
			CREATE INDEX IF NOT EXISTS idx_import_logs_status ON import_logs(status);
			CREATE INDEX IF NOT EXISTS idx_import_logs_imported_at ON import_logs(imported_at);
		`)
		if err != nil {
			return fmt.Errorf("failed to create import_logs table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] drop import_logs table")

		_, err := db.ExecContext(ctx, `
			DROP TABLE IF EXISTS import_logs CASCADE
		`)
		if err != nil {
			return fmt.Errorf("failed to drop import_logs table: %w", err)
		}

		return nil
	})
}
