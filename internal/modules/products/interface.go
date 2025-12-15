package products

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/uptrace/bun"
)

type ProductService interface {
	// Regular methods (without transaction)
	GetProductByID(ctx context.Context, id int) (*model.Product, error)
	GetProductByName(ctx context.Context, name string) (*model.Product, error)
	GetOrCreateProduct(ctx context.Context, product *model.Product) (*model.Product, error)

	// Transaction methods (for batch import)
	GetOrCreateProductInTx(ctx context.Context, tx *bun.Tx, product *model.Product) (*model.Product, error)
}

type VariantService interface {
	// Regular methods (without transaction)
	GetVariantByID(ctx context.Context, id int) (*model.Variant, error)
	GetVariantByProductIDAndName(ctx context.Context, productID int, name string) (*model.Variant, error)
	GetOrCreateVariant(ctx context.Context, variant *model.Variant) (*model.Variant, error)

	// Transaction methods (for batch import)
	GetOrCreateVariantInTx(ctx context.Context, tx *bun.Tx, variant *model.Variant) (*model.Variant, error)
}
