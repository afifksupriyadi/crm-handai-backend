// internal/modules/importdata/parser/product_parser.go

package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/constant"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
)

var (
	sizeNotationRegex = regexp.MustCompile(`\s*(250ml|500ml|1L|1l)\s*`)
	multiSpaceRegex   = regexp.MustCompile(`\s+`)
)

// ParseProduct extracts structured product information from Excel product name
func ParseProduct(excelName string) (*model.ParsedProductInfo, error) {
	excelName = strings.TrimSpace(excelName)

	// Normalize: remove "(Medium)" - cosmetic only
	normalizedName := strings.ReplaceAll(excelName, " (Medium)", "")

	// Normalize: remove size notation (250ml, 500ml, 1L) from product name
	// This handles cases like "Americano 250ml (Botol)" → "Americano (Botol)"
	normalizedName = sizeNotationRegex.ReplaceAllString(normalizedName, " ")
	normalizedName = strings.TrimSpace(normalizedName)
	normalizedName = multiSpaceRegex.ReplaceAllString(normalizedName, " ")

	// Normalize case variations: (botol) → (Botol)
	normalizedName = strings.ReplaceAll(normalizedName, "(botol)", "(Botol)")

	// Lookup product info from mapping
	productInfo, exists := constant.ProductMapping[normalizedName]
	if !exists {
		return nil, fmt.Errorf("product not found in mapping: %s (normalized: %s)", excelName, normalizedName)
	}

	return &model.ParsedProductInfo{
		OriginalName:   excelName,
		NormalizedName: productInfo.Name,
		Category:       productInfo.Category.String(),
		BasePrice:      productInfo.BasePrice,
		HasVariants:    productInfo.HasVariants,
	}, nil
}
