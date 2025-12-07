package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	phoneRegex = regexp.MustCompile(`^0[0-9]{8,13}$`)
)

// NormalizeCustomer normalizes customer name and phone from Excel
func NormalizeCustomer(name string, phone string) (*model.ParsedCustomerInfo, error) {
	normalizedName := NormalizeName(name)
	normalizedPhone, err := NormalizePhone(phone)
	if err != nil {
		return nil, err
	}

	return &model.ParsedCustomerInfo{
		Name:  normalizedName,
		Phone: normalizedPhone,
	}, nil
}

// NormalizeName normalizes customer name (title case, trim spaces)
func NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Join(strings.Fields(name), " ")
	return toTitleCase(name)
}

// NormalizePhone normalizes phone number (remove spaces, standardize format)
func NormalizePhone(phone string) (string, error) {
	// Clean phone number
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.TrimSpace(phone)
	phone = strings.TrimPrefix(phone, "+")

	// Check if empty after cleaning
	if phone == "" {
		return "", fmt.Errorf("phone number is empty or invalid")
	}

	// Convert 62xxx to 0xxx (Indonesian format)
	if strings.HasPrefix(phone, "62") {
		phone = "0" + phone[2:]
	}

	// Validate
	if !phoneRegex.MatchString(phone) {
		return "", fmt.Errorf("invalid phone format '%s' (must be 0xxx, 9-14 digits)", phone)
	}

	return phone, nil
}

// toTitleCase converts string to title case using Indonesian language rules
func toTitleCase(s string) string {
	caser := cases.Title(language.Indonesian)
	return caser.String(strings.ToLower(s))
}
