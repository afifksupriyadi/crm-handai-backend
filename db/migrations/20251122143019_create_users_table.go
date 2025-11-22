package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create users table")

		query := `
			CREATE TABLE users (
				id SERIAL PRIMARY KEY,
				email VARCHAR(100) NOT NULL UNIQUE,
				name VARCHAR(255) NOT NULL,
				password_hash TEXT NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				created_by INTEGER,
				updated_at TIMESTAMPTZ,
				updated_by INTEGER,
				CONSTRAINT users_status_check CHECK (status IN ('ACTIVE', 'INACTIVE')),
				CONSTRAINT fk_users_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
				CONSTRAINT fk_users_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
			);

			CREATE INDEX idx_users_email ON users(email);
			CREATE INDEX idx_users_status ON users(status);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create users table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop users table")

		query := `
			DROP INDEX IF EXISTS idx_users_email;
			DROP INDEX IF EXISTS idx_users_status;
			DROP TABLE IF EXISTS users;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop users table: %w", err)
		}

		return nil
	})
}
