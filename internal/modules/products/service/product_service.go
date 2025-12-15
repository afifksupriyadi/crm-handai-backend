package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

type ProductServiceImpl struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) products.ProductService {
	return &ProductServiceImpl{
		productRepo: productRepo,
	}
}

// ==========================================
// REGULAR METHODS (Without Transaction)
// ==========================================

func (s *ProductServiceImpl) GetProductByID(ctx context.Context, id int) (*model.Product, error) {
	product, err := s.productRepo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrProductNotFound, "Product not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get product")
	}

	return product, nil
}

func (s *ProductServiceImpl) GetProductByName(ctx context.Context, name string) (*model.Product, error) {
	product, err := s.productRepo.GetProductByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrProductNotFound, "Product not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get product")
	}

	return product, nil
}

func (s *ProductServiceImpl) GetOrCreateProduct(ctx context.Context, product *model.Product) (*model.Product, error) {
	// Try to get existing product
	existing, err := s.productRepo.GetProductByName(ctx, product.Name)

	// If found, return it
	if err == nil {
		return existing, nil
	}

	// If not found, create new
	if errors.Is(err, sql.ErrNoRows) {
		err = s.productRepo.CreateProduct(ctx, product)
		if err != nil {
			return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create product")
		}

		logger.FromContext(ctx, 1).Info().
			Int("product_id", product.ID).
			Str("name", product.Name).
			Msg("Product created")

		return product, nil
	}

	// Other database error
	return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check existing product")
}

// ==========================================
// TRANSACTION METHODS (With Transaction)
// ==========================================

func (s *ProductServiceImpl) GetOrCreateProductInTx(ctx context.Context, tx *bun.Tx, product *model.Product) (*model.Product, error) {
	// Try to get existing product
	existing, err := s.productRepo.GetProductByNameInTx(ctx, tx, product.Name)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check existing product")
	}

	// If found, return it
	if existing != nil {
		return existing, nil
	}

	// If not found, create new
	err = s.productRepo.CreateProductInTx(ctx, tx, product)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create product")
	}

	logger.FromContext(ctx, 1).Debug().
		Int("product_id", product.ID).
		Str("name", product.Name).
		Msg("Product created in transaction")

	return product, nil
}
