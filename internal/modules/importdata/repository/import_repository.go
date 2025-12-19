package repository

import (
	"github.com/uptrace/bun"
)

type ImportRepository interface {
	// Reserved for future use if needed
}

type ImportRepositoryImpl struct {
	db *bun.DB
}

// NewImportRepository creates a new instance of ImportRepositoryImpl
func NewImportRepository(db *bun.DB) ImportRepository {
	return &ImportRepositoryImpl{db: db}
}
