// internal/modules/importdata/parser/payment_parser.go

package parser

import (
	"fmt"
	"strings"
)

// NormalizePaymentMethod normalizes payment method strings from Excel to match database constraints
// Database constraint: CHECK (payment_method IN ('Tunai', 'Non Tunai'))
//
// Handles cases where Excel contains multiple payment methods separated by comma
// (e.g., "Tunai, Non Tunai" when multiple products in same transaction)
// We take the FIRST payment method since all products in a transaction should use same payment method
func NormalizePaymentMethod(input string) (string, error) {
	// Trim all whitespace (including leading/trailing and multiple spaces)
	normalized := strings.TrimSpace(input)

	// If comma-separated, take the first value only
	// Example: "Tunai, Non Tunai" → "Tunai"
	if strings.Contains(normalized, ",") {
		parts := strings.Split(normalized, ",")
		normalized = strings.TrimSpace(parts[0])
	}

	// Convert to lowercase for case-insensitive comparison
	lower := strings.ToLower(normalized)

	// Map common variations to standard values
	switch lower {
	case "tunai", "cash", "uang tunai":
		return "Tunai", nil
	case "non tunai", "nontunai", "non-tunai", "transfer", "cashless", "e-wallet", "digital":
		return "Non Tunai", nil
	default:
		// If it doesn't match known patterns, return error with the actual value
		return "", fmt.Errorf("invalid payment method: '%s' (must be 'Tunai' or 'Non Tunai')", input)
	}
}
