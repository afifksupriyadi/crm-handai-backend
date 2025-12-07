package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/uptrace/bun"
)

type ProductRepository interface {
	GetProductByID(ctx context.Context, id int) (*model.Product, error)
	GetProductByName(ctx context.Context, name string) (*model.Product, error)
	CreateProduct(ctx context.Context, product *model.Product) error
	GetOrCreateProduct(ctx context.Context, product *model.Product) (*model.Product, error)
	UpdateProduct(ctx context.Context, product *model.Product) error
}

type ProductRepositoryImpl struct {
	db *bun.DB
}

func NewProductRepository(db *bun.DB) ProductRepository {
	return &ProductRepositoryImpl{db: db}
}

func (r *ProductRepositoryImpl) GetProductByID(ctx context.Context, id int) (*model.Product, error) {
	product := new(model.Product)
	err := r.db.NewSelect().
		Model(product).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	return product, nil
}

func (r *ProductRepositoryImpl) GetProductByName(ctx context.Context, name string) (*model.Product, error) {
	product := new(model.Product)
	err := r.db.NewSelect().
		Model(product).
		Where("name = ?", name).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	return product, nil
}

func (r *ProductRepositoryImpl) CreateProduct(ctx context.Context, product *model.Product) error {
	_, err := r.db.NewInsert().
		Model(product).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (r *ProductRepositoryImpl) GetOrCreateProduct(ctx context.Context, product *model.Product) (*model.Product, error) {
	existing, err := r.GetProductByName(ctx, product.Name)
	if err == nil {
		return existing, nil
	}

	err = r.CreateProduct(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (r *ProductRepositoryImpl) UpdateProduct(ctx context.Context, product *model.Product) error {
	_, err := r.db.NewUpdate().
		Model(product).
		Where("id = ?", product.ID).
		Where("deleted_at IS NULL").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}
