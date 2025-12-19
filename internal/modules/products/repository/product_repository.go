package repository

import (
	"context"
	"database/sql"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

type ProductRepository interface {
	GetProductByID(ctx context.Context, id int) (*model.Product, error)
	GetProductByName(ctx context.Context, name string) (*model.Product, error)
	CreateProduct(ctx context.Context, product *model.Product) error
	GetProductByNameInTx(ctx context.Context, tx *bun.Tx, name string) (*model.Product, error)
	CreateProductInTx(ctx context.Context, tx *bun.Tx, product *model.Product) error
}

type ProductRepositoryImpl struct {
	db *bun.DB
}

func NewProductRepository(db *bun.DB) ProductRepository {
	return &ProductRepositoryImpl{db: db}
}

// Regular methods
func (r *ProductRepositoryImpl) GetProductByID(ctx context.Context, id int) (*model.Product, error) {
	return r.getProductByID(ctx, r.db, id)
}

func (r *ProductRepositoryImpl) GetProductByName(ctx context.Context, name string) (*model.Product, error) {
	return r.getProductByName(ctx, r.db, name)
}

func (r *ProductRepositoryImpl) CreateProduct(ctx context.Context, product *model.Product) error {
	return r.createProduct(ctx, r.db, product)
}

// Transaction methods
func (r *ProductRepositoryImpl) GetProductByNameInTx(ctx context.Context, tx *bun.Tx, name string) (*model.Product, error) {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}
	return r.getProductByName(ctx, db, name)
}

func (r *ProductRepositoryImpl) CreateProductInTx(ctx context.Context, tx *bun.Tx, product *model.Product) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}
	return r.createProduct(ctx, db, product)
}

// Internal shared implementations
func (r *ProductRepositoryImpl) getProductByID(ctx context.Context, db bun.IDB, id int) (*model.Product, error) {
	product := new(model.Product)
	err := db.NewSelect().Model(product).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find product by ID")
		return nil, err
	}
	return product, nil
}

func (r *ProductRepositoryImpl) getProductByName(ctx context.Context, db bun.IDB, name string) (*model.Product, error) {
	product := new(model.Product)
	err := db.NewSelect().Model(product).Where("name = ?", name).Where("deleted_at IS NULL").Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find product by name")
		return nil, err
	}
	return product, nil
}

func (r *ProductRepositoryImpl) createProduct(ctx context.Context, db bun.IDB, product *model.Product) error {
	_, err := db.NewInsert().Model(product).Exec(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create product")
		return err
	}
	return nil
}
