package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/constant"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
)

var (
	sizeInNameRegex = regexp.MustCompile(`(250ml|500ml|1L|1l)`)
)

// ParseVariant extracts variant info from Excel data
// PRIORITY: Variant Column > Base Price Analysis > Size in Name > Default
func ParseVariant(productName string, variantText string, hargaVarianTotal float64, quantity int) (*model.ParsedVariantInfo, error) {
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

	variantText = strings.TrimSpace(variantText)

	// PRIORITY 1: Variant column is provided (most accurate)
	if variantText != "" {
		return validateVariantFromExcel(parsed.NormalizedName, productName, variantText, hargaVarianTotal, quantity)
	}

	// PRIORITY 2: Variant column empty - check if product name has size notation
	if sizeMatch := sizeInNameRegex.FindString(productName); sizeMatch != "" {
		normalizedSize := normalizeSizeNotation(sizeMatch)
		return getVariantFromSize(parsed.NormalizedName, normalizedSize)
	}

	// PRIORITY 3: No variant column, no size in name - use default
	return getDefaultVariantFromRules(parsed.NormalizedName)
}

// normalizeSizeNotation converts size notation to standard format
func normalizeSizeNotation(size string) string {
	size = strings.ToLower(strings.TrimSpace(size))

	if size == "1l" {
		return "1000ml"
	}

	return size
}

// getVariantFromSize looks up variant info by size
func getVariantFromSize(productName string, size string) (*model.ParsedVariantInfo, error) {
	rules, exists := constant.VariantPricingRules[productName]
	if !exists {
		return nil, fmt.Errorf("variant pricing rules not found for product: %s", productName)
	}

	variantSize := constant.VariantSize(size)

	for _, rule := range rules {
		if rule.Size == variantSize {
			return &model.ParsedVariantInfo{
				Size:          rule.Size.String(),
				PriceModifier: rule.PriceModifier,
				IsDefault:     rule.IsDefault,
			}, nil
		}
	}

	return nil, fmt.Errorf("variant size %s not found in rules for product %s", size, productName)
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
func validateVariantFromExcel(normalizedName, originalName, variantText string, hargaVarianTotal float64, quantity int) (*model.ParsedVariantInfo, error) {
	if quantity == 0 {
		return nil, fmt.Errorf("quantity cannot be zero")
	}

	priceModifierPerUnit := hargaVarianTotal / float64(quantity)
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
					"price mismatch for %s variant %s: expected %f, got %f",
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
