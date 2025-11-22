package migrations

import (
	"context"
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/security"
	"github.com/uptrace/bun"
)

const (
	DefaultAdminUsername = "admin"
	DefaultAdminPassword = "admin123"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] seed admin user")

		hashedPassword, err := security.HashPassword(DefaultAdminPassword)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		query := `
			INSERT INTO users (username, password)
			VALUES (?, ?)
			ON CONFLICT (username) DO NOTHING
		`
		_, err = db.ExecContext(ctx, query, DefaultAdminUsername, hashedPassword)
		if err != nil {
			return fmt.Errorf("failed to seed admin user: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] remove admin user")

		query := `DELETE FROM users WHERE username = ?`

		_, err := db.ExecContext(ctx, query, DefaultAdminUsername)
		if err != nil {
			return fmt.Errorf("failed to remove admin user: %w", err)
		}

		return nil
	})
}
