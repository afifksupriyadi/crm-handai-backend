package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Customer struct {
	bun.BaseModel `bun:"table:customers"`

	ID        int        `bun:"id,pk,autoincrement"`
	Name      string     `bun:"name,notnull"`
	Phone     string     `bun:"phone,notnull,unique"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt *time.Time `bun:"updated_at,nullzero"`
	DeletedAt *time.Time `bun:"deleted_at,soft_delete,nullzero"`
}
