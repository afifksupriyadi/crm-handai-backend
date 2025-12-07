package parser

import (
	"fmt"
	"strings"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/constant"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
)

// ParseVariant extracts variant info from Excel data
func ParseVariant(productName string, variantText string, hargaVarianTotal int64, quantity int) (*model.ParsedVariantInfo, error) {
	parsed, err := ParseProduct(productName)
	if err != nil {
		return nil, err
	}

	// Product without variants - return default
	if !parsed.HasVariants {
		return &model.ParsedVariantInfo{
			Size:          constant.VariantSize250ml.String(),
			PriceModifier: 0,
			IsDefault:     true,
		}, nil
	}

	// Empty variant text - get default from rules
	variantText = strings.TrimSpace(variantText)
	if variantText == "" {
		return getDefaultVariantFromRules(parsed.NormalizedName)
	}

	// Validate variant from Excel
	return validateVariantFromExcel(parsed.NormalizedName, productName, variantText, hargaVarianTotal, quantity)
}

// getDefaultVariantFromRules retrieves default variant for a product
func getDefaultVariantFromRules(productName string) (*model.ParsedVariantInfo, error) {
	rules, exists := constant.VariantPricingRules[productName]
	if !exists {
		return nil, fmt.Errorf("variant pricing rules not found for product: %s", productName)
	}

	for _, rule := range rules {
		if rule.IsDefault {
			return &model.ParsedVariantInfo{
				Size:          rule.Size.String(),
				PriceModifier: rule.PriceModifier,
				IsDefault:     true,
			}, nil
		}
	}

	return nil, fmt.Errorf("no default variant found for product: %s", productName)
}

// validateVariantFromExcel validates variant data from Excel against pricing rules
func validateVariantFromExcel(normalizedName, originalName, variantText string, hargaVarianTotal int64, quantity int) (*model.ParsedVariantInfo, error) {
	if quantity == 0 {
		return nil, fmt.Errorf("quantity cannot be zero")
	}

	priceModifierPerUnit := hargaVarianTotal / int64(quantity)
	variantSize := constant.VariantSize(variantText)

	rules, exists := constant.VariantPricingRules[normalizedName]
	if !exists {
		return nil, fmt.Errorf("variant pricing rules not found for product: %s", normalizedName)
	}

	for _, rule := range rules {
		if rule.Size == variantSize {
			// Validate price modifier matches
			if rule.PriceModifier != priceModifierPerUnit {
				return nil, fmt.Errorf(
					"price mismatch for %s variant %s: expected %d, got %d",
					originalName, variantSize, rule.PriceModifier, priceModifierPerUnit,
				)
			}

			return &model.ParsedVariantInfo{
				Size:          rule.Size.String(),
				PriceModifier: priceModifierPerUnit,
				IsDefault:     rule.IsDefault,
			}, nil
		}
	}

	return nil, fmt.Errorf("variant %s not found in rules for product %s", variantSize, originalName)
}
