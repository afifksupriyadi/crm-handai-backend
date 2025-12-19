package repository

import (
	"context"
	"database/sql"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

type VariantRepository interface {
	GetVariantByID(ctx context.Context, id int) (*model.Variant, error)
	GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error)
	CreateVariant(ctx context.Context, variant *model.Variant) error
	GetVariantByProductIDAndNameInTx(ctx context.Context, tx *bun.Tx, productID int, name string) (*model.Variant, error)
	CreateVariantInTx(ctx context.Context, tx *bun.Tx, variant *model.Variant) error
}

type VariantRepositoryImpl struct {
	db *bun.DB
}

func NewVariantRepository(db *bun.DB) VariantRepository {
	return &VariantRepositoryImpl{db: db}
}

// Regular methods
func (r *VariantRepositoryImpl) GetVariantByID(ctx context.Context, id int) (*model.Variant, error) {
	return r.getVariantByID(ctx, r.db, id)
}

func (r *VariantRepositoryImpl) GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error) {
	return r.getVariantByProductIDAndName(ctx, r.db, productID, name)
}

func (r *VariantRepositoryImpl) CreateVariant(ctx context.Context, variant *model.Variant) error {
	return r.createVariant(ctx, r.db, variant)
}

// Transaction methods
func (r *VariantRepositoryImpl) GetVariantByProductIDAndNameInTx(ctx context.Context, tx *bun.Tx, productID int, name string) (*model.Variant, error) {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}
	return r.getVariantByProductIDAndName(ctx, db, productID, name)
}

func (r *VariantRepositoryImpl) CreateVariantInTx(ctx context.Context, tx *bun.Tx, variant *model.Variant) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}
	return r.createVariant(ctx, db, variant)
}

// Internal shared implementations
func (r *VariantRepositoryImpl) getVariantByID(ctx context.Context, db bun.IDB, id int) (*model.Variant, error) {
	variant := new(model.Variant)
	err := db.NewSelect().Model(variant).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find variant by ID")
		return nil, err
	}
	return variant, nil
}

func (r *VariantRepositoryImpl) getVariantByProductIDAndName(ctx context.Context, db bun.IDB, productID int, name string) (*model.Variant, error) {
	variant := new(model.Variant)
	err := db.NewSelect().Model(variant).Where("product_id = ?", productID).Where("name = ?", name).Where("deleted_at IS NULL").Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find variant")
		return nil, err
	}
	return variant, nil
}

func (r *VariantRepositoryImpl) createVariant(ctx context.Context, db bun.IDB, variant *model.Variant) error {
	_, err := db.NewInsert().Model(variant).Exec(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create variant")
		return err
	}
	return nil
}
