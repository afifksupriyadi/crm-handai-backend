package parser

import (
	"fmt"
	"time"
)

// ParseTransactionDateWithTimezone parses transaction date and time from Excel
func ParseTransactionDateWithTimezone(tanggal, jam string) (time.Time, error) {
	// Combine date and time
	datetime := fmt.Sprintf("%s %s", tanggal, jam)

	// Parse as naive datetime first
	parsedTime, err := time.Parse("02-01-2006 15:04:05", datetime)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse transaction date: %w", err)
	}

	// Load Asia/Jakarta timezone
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to load WIB timezone: %w", err)
	}

	// Reconstruct time with WIB timezone
	parsedTimeWIB := time.Date(
		parsedTime.Year(),
		parsedTime.Month(),
		parsedTime.Day(),
		parsedTime.Hour(),
		parsedTime.Minute(),
		parsedTime.Second(),
		0,
		loc, // Apply WIB timezone
	)

	return parsedTimeWIB, nil
}
