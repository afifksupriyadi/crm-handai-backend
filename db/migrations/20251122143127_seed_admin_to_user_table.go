package migrations

import (
	"context"
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/security"
	"github.com/uptrace/bun"
)

const (
	DefaultAdminEmail    = "admin@handai.com"
	DefaultAdminName     = "Administrator"
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
			INSERT INTO users (email, name, password_hash, status, created_at)
			VALUES (?, ?, ?, 'ACTIVE', CURRENT_TIMESTAMP)
			ON CONFLICT (email) DO NOTHING
		`

		_, err = db.ExecContext(ctx, query, DefaultAdminEmail, DefaultAdminName, hashedPassword)
		if err != nil {
			return fmt.Errorf("failed to seed admin user: %w", err)
		}

		fmt.Printf(" [seed] Default admin: %s / %s\n", DefaultAdminEmail, DefaultAdminPassword)

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] remove admin user")

		query := `DELETE FROM users WHERE email = ?`

		_, err := db.ExecContext(ctx, query, DefaultAdminEmail)
		if err != nil {
			return fmt.Errorf("failed to remove admin user: %w", err)
		}

		return nil
	})
}
