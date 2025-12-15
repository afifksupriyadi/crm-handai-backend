package repository

import (
	"context"
	"database/sql"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/uptrace/bun"
)

type VariantRepository interface {
	// Regular methods (without transaction)
	GetVariantByID(ctx context.Context, id int) (*model.Variant, error)
	GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error)
	CreateVariant(ctx context.Context, variant *model.Variant) error

	// Transaction methods (for batch import)
	GetVariantByProductIDAndNameInTx(ctx context.Context, tx *bun.Tx, productID int, name string) (*model.Variant, error)
	CreateVariantInTx(ctx context.Context, tx *bun.Tx, variant *model.Variant) error
}

type VariantRepositoryImpl struct {
	db *bun.DB
}

func NewVariantRepository(db *bun.DB) VariantRepository {
	return &VariantRepositoryImpl{db: db}
}

// ==========================================
// REGULAR METHODS (Without Transaction)
// ==========================================

func (r *VariantRepositoryImpl) GetVariantByID(ctx context.Context, id int) (*model.Variant, error) {
	variant := new(model.Variant)
	err := r.db.NewSelect().
		Model(variant).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return variant, nil
}

func (r *VariantRepositoryImpl) GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error) {
	variant := new(model.Variant)
	err := r.db.NewSelect().
		Model(variant).
		Where("product_id = ?", productID).
		Where("name = ?", name).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return variant, nil
}

func (r *VariantRepositoryImpl) CreateVariant(ctx context.Context, variant *model.Variant) error {
	_, err := r.db.NewInsert().
		Model(variant).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ==========================================
// TRANSACTION METHODS (With Transaction)
// ==========================================

func (r *VariantRepositoryImpl) GetVariantByProductIDAndNameInTx(ctx context.Context, tx *bun.Tx, productID int, name string) (*model.Variant, error) {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	variant := new(model.Variant)
	err := db.NewSelect().
		Model(variant).
		Where("product_id = ?", productID).
		Where("name = ?", name).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}

	if err != nil {
		return nil, err
	}

	return variant, nil
}

func (r *VariantRepositoryImpl) CreateVariantInTx(ctx context.Context, tx *bun.Tx, variant *model.Variant) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	_, err := db.NewInsert().
		Model(variant).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
