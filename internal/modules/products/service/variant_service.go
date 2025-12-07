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

type VariantServiceImpl struct {
	variantRepo repository.VariantRepository
}

func NewVariantService(variantRepo repository.VariantRepository) products.VariantService {
	return &VariantServiceImpl{
		variantRepo: variantRepo,
	}
}

func (s *VariantServiceImpl) GetVariantByID(ctx context.Context, id int) (*model.Variant, error) {
	variant, err := s.variantRepo.GetVariantByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrVariantNotFound, "Variant not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get variant")
	}

	return variant, nil
}

func (s *VariantServiceImpl) GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error) {
	variant, err := s.variantRepo.GetVariantByProductIDAndName(ctx, productID, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrVariantNotFound, "Variant not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get variant")
	}

	return variant, nil
}

func (s *VariantServiceImpl) GetOrCreateVariant(ctx context.Context, variant *model.Variant) (*model.Variant, error) {
	// Try to get existing variant
	existing, err := s.variantRepo.GetVariantByProductIDAndName(ctx, variant.ProductID, variant.Name)

	// If found, return it
	if err == nil {
		return existing, nil
	}

	// If not found, create new
	if errors.Is(err, sql.ErrNoRows) {
		err = s.variantRepo.CreateVariant(ctx, variant)
		if err != nil {
			return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create variant")
		}
		return variant, nil
	}

	// Other database error
	return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check existing variant")
}
