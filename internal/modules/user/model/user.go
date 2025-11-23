package model

import (
	"time"

	"github.com/uptrace/bun"
)

// User represents a user in database
type User struct {
	bun.BaseModel `bun:"table:users"`

	ID           int       `bun:"id,pk,autoincrement"`
	Email        string    `bun:"email,notnull,unique"`
	Name         string    `bun:"name,notnull"`
	PasswordHash string    `bun:"password_hash,notnull"`
	Status       string    `bun:"status,default:ACTIVE,nullzero"`
	CreatedAt    time.Time `bun:"created_at,default:current_timestamp,nullzero"`
	CreatedBy    int       `bun:"created_by,nullzero"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero"`
	UpdatedBy    int       `bun:"updated_by,nullzero"`
}
