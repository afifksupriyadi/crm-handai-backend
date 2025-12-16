package model

// AuthUser represents a user for authentication purposes.
type AuthUser struct {
	ID           int    `bun:"id"`
	Name         string `bun:"name"`
	Email        string `bun:"email"`
	PasswordHash string `bun:"password_hash"`
	Status       string `bun:"status"`
}
