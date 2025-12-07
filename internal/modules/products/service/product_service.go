// internal/modules/products/service/product_service.go

package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

type ProductServiceImpl struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) products.ProductService {
	return &ProductServiceImpl{
		productRepo: productRepo,
	}
}

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
	existing, err := s.productRepo.GetProductByName(ctx, product.Name)
	if err == nil {
		return existing, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check existing product")
	}

	err = s.productRepo.CreateProduct(ctx, product)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create product")
	}

	return product, nil
}
