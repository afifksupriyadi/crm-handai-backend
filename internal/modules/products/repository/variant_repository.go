// internal/modules/products/repository/variant_repository.go

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/uptrace/bun"
)

type VariantRepository interface {
	GetVariantByID(ctx context.Context, id int) (*model.Variant, error)
	GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error)
	CreateVariant(ctx context.Context, variant *model.Variant) error
	GetOrCreateVariant(ctx context.Context, variant *model.Variant) (*model.Variant, error)
	UpdateVariant(ctx context.Context, variant *model.Variant) error
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("variant not found")
		}
		return nil, fmt.Errorf("failed to get variant: %w", err)
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("variant not found")
		}
		return nil, fmt.Errorf("failed to get variant: %w", err)
	}

	return variant, nil
}

func (r *VariantRepositoryImpl) CreateVariant(ctx context.Context, variant *model.Variant) error {
	_, err := r.db.NewInsert().
		Model(variant).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create variant: %w", err)
	}

	return nil
}

func (r *VariantRepositoryImpl) GetOrCreateVariant(ctx context.Context, variant *model.Variant) (*model.Variant, error) {
	existing, err := r.GetVariantByProductIDAndName(ctx, variant.ProductID, variant.Name)
	if err == nil {
		return existing, nil
	}

	err = r.CreateVariant(ctx, variant)
	if err != nil {
		return nil, fmt.Errorf("failed to create variant: %w", err)
	}

	return variant, nil
}

func (r *VariantRepositoryImpl) UpdateVariant(ctx context.Context, variant *model.Variant) error {
	_, err := r.db.NewUpdate().
		Model(variant).
		Where("id = ?", variant.ID).
		Where("deleted_at IS NULL").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update variant: %w", err)
	}

	return nil
}
