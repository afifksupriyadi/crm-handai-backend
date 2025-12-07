package products

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
)

type ProductService interface {
	GetProductByID(ctx context.Context, id int) (*model.Product, error)
	GetProductByName(ctx context.Context, name string) (*model.Product, error)
	GetOrCreateProduct(ctx context.Context, product *model.Product) (*model.Product, error)
}

type VariantService interface {
	GetVariantByID(ctx context.Context, id int) (*model.Variant, error)
	GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error)
	GetOrCreateVariant(ctx context.Context, variant *model.Variant) (*model.Variant, error)
}
