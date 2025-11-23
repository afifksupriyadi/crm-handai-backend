package repository

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/user/model"
	"github.com/uptrace/bun"
)

// UserRepository defines the interface for user-related database operations.
type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*model.AuthUser, error)
	UpdateUser(ctx context.Context, user *model.User) error
}

// UserRepositoryImpl provides the concrete implementation of UserRepository using bun.
type UserRepositoryImpl struct {
	DB bun.IDB
}

// NewUserRepository returns a new instance of UserRepositoryImpl.
func NewUserRepository(db bun.IDB) UserRepository {
	return &UserRepositoryImpl{DB: db}
}

// GetUserByEmail retrieves a user by their email address.
// Used for authentication purposes.
func (r *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*model.AuthUser, error) {
	var user model.AuthUser

	err := r.DB.NewSelect().
		Table("users").
		ColumnExpr("users.id").
		ColumnExpr("users.name").
		ColumnExpr("users.status").
		ColumnExpr("users.password_hash").
		Where("users.email = ?", email).
		Scan(ctx, &user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update user updates an existing user in the database.
// Used for changes password purposes.
func (r *UserRepositoryImpl) UpdateUser(ctx context.Context, user *model.User) error {
	// TODO: implement update user for password changes
	return nil
}
