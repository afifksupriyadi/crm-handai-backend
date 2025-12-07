package repository

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/uptrace/bun"
)

type VariantRepository interface {
	GetVariantByID(ctx context.Context, id int) (*model.Variant, error)
	GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error)
	CreateVariant(ctx context.Context, variant *model.Variant) error
}

type VariantRepositoryImpl struct {
	db *bun.DB
}

func NewVariantRepository(db *bun.DB) VariantRepository {
	return &VariantRepositoryImpl{db: db}
}

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
