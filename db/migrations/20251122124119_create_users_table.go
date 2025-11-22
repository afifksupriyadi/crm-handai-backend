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
				username VARCHAR(20) NOT NULL UNIQUE,
				password VARCHAR(255) NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ,
				deleted_at TIMESTAMPTZ
			);

			CREATE INDEX idx_users_username ON users(username);
			CREATE INDEX idx_users_deleted_at ON users(deleted_at);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create users table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop users table")

		query := `
			DROP INDEX IF EXISTS idx_users_username;
			DROP INDEX IF EXISTS idx_users_deleted_at;
			DROP TABLE IF EXISTS users;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop users table: %w", err)
		}

		return nil
	})
}
