package repository

import (
	"context"
	"database/sql"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/uptrace/bun"
)

type ProductRepository interface {
	// Regular methods (without transaction)
	GetProductByID(ctx context.Context, id int) (*model.Product, error)
	GetProductByName(ctx context.Context, name string) (*model.Product, error)
	CreateProduct(ctx context.Context, product *model.Product) error

	// Transaction methods (for batch import)
	GetProductByNameInTx(ctx context.Context, tx *bun.Tx, name string) (*model.Product, error)
	CreateProductInTx(ctx context.Context, tx *bun.Tx, product *model.Product) error
}

type ProductRepositoryImpl struct {
	db *bun.DB
}

func NewProductRepository(db *bun.DB) ProductRepository {
	return &ProductRepositoryImpl{db: db}
}

// ==========================================
// REGULAR METHODS (Without Transaction)
// ==========================================

func (r *ProductRepositoryImpl) GetProductByID(ctx context.Context, id int) (*model.Product, error) {
	product := new(model.Product)
	err := r.db.NewSelect().
		Model(product).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return product, nil
}

func (r *ProductRepositoryImpl) CreateProduct(ctx context.Context, product *model.Product) error {
	_, err := r.db.NewInsert().
		Model(product).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ==========================================
// TRANSACTION METHODS (With Transaction)
// ==========================================

func (r *ProductRepositoryImpl) GetProductByNameInTx(ctx context.Context, tx *bun.Tx, name string) (*model.Product, error) {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	product := new(model.Product)
	err := db.NewSelect().
		Model(product).
		Where("name = ?", name).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *ProductRepositoryImpl) CreateProductInTx(ctx context.Context, tx *bun.Tx, product *model.Product) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	_, err := db.NewInsert().
		Model(product).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
