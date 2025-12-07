package parser

import (
	"fmt"
	"strings"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/constant"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
)

// ParseProduct extracts structured product information from Excel product name
func ParseProduct(excelName string) (*model.ParsedProductInfo, error) {
	excelName = strings.TrimSpace(excelName)

	// Normalize: remove "(Medium)" - it's cosmetic only
	normalizedName := strings.ReplaceAll(excelName, " (Medium)", "")
	normalizedName = strings.TrimSpace(normalizedName)

	// Lookup product info from mapping
	productInfo, exists := constant.ProductMapping[normalizedName]
	if !exists {
		return nil, fmt.Errorf("product not found in mapping: %s", excelName)
	}

	return &model.ParsedProductInfo{
		OriginalName:   excelName,
		NormalizedName: productInfo.Name,
		Category:       productInfo.Category.String(),
		BasePrice:      productInfo.BasePrice,
		HasVariants:    productInfo.HasVariants,
	}, nil
}
