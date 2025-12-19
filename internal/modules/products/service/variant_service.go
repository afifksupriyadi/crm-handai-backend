package service

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

type VariantServiceImpl struct {
	variantRepo repository.VariantRepository // ✅ ONLY repo!
}

func NewVariantService(variantRepo repository.VariantRepository) products.VariantService {
	return &VariantServiceImpl{
		variantRepo: variantRepo,
	}
}

func (s *VariantServiceImpl) GetVariantByID(ctx context.Context, id int) (*model.Variant, error) {
	variant, err := s.variantRepo.GetVariantByID(ctx, id)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get variant")
	}
	if variant == nil {
		return nil, response.WrapAppError(ctx, nil, response.ErrVariantNotFound, "Variant not found")
	}
	return variant, nil
}

func (s *VariantServiceImpl) GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error) {
	variant, err := s.variantRepo.GetVariantByProductIDAndName(ctx, productID, name)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get variant")
	}
	if variant == nil {
		return nil, response.WrapAppError(ctx, nil, response.ErrVariantNotFound, "Variant not found")
	}
	return variant, nil
}

func (s *VariantServiceImpl) GetOrCreateVariant(ctx context.Context, variant *model.Variant) (*model.Variant, error) {
	existing, err := s.variantRepo.GetVariantByProductIDAndName(ctx, variant.ProductID, variant.Name)
	if err == nil {
		return existing, nil
	}

	err = s.variantRepo.CreateVariant(ctx, variant)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create variant")
	}

	logger.FromContext(ctx, 1).Info().Int("variant_id", variant.ID).Int("product_id", variant.ProductID).Str("name", variant.Name).Msg("Variant created")
	return variant, nil
}

func (s *VariantServiceImpl) GetOrCreateVariantInTx(ctx context.Context, tx *bun.Tx, variant *model.Variant) (*model.Variant, error) {
	existing, err := s.variantRepo.GetVariantByProductIDAndNameInTx(ctx, tx, variant.ProductID, variant.Name)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check existing variant")
	}

	if existing != nil {
		return existing, nil
	}

	err = s.variantRepo.CreateVariantInTx(ctx, tx, variant)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create variant")
	}

	logger.FromContext(ctx, 1).Debug().Int("variant_id", variant.ID).Int("product_id", variant.ProductID).Str("name", variant.Name).Msg("Variant created in transaction")
	return variant, nil
}
