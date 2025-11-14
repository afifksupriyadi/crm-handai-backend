package db

import (
	"fmt"
	"strings"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/uptrace/bun"
)

// BunFactory represents a database connection factory that can handle specific database URI prefixes.
type BunFactory struct {
	Prefixes []string
	Opener   func(c *config.Config) (*bun.DB, error)
}

var bunFactories []*BunFactory = []*BunFactory{}

func allPrefixes() string {
	// Pre-allocate slice with known capacity.
	prefixes := make([]string, 0, len(bunFactories)*2)
	for _, bunFactory := range bunFactories {
		prefixes = append(prefixes, bunFactory.Prefixes...)
	}
	return strings.Join(prefixes, "|")
}

// Open establishes a database connection based on the provided configuration.
// It supports multiple database types through the BunFactory pattern, where each factory
// handles specific database URI prefixes (e.g., postgres://, postgresql://, unix://).
//
// The function will:
//   - Return nil if no database URI is provided
//   - Attempt to match the URI against registered database factories
//   - Configure OpenTelemetry and debug hooks for the connection
//   - Return an error if the URI prefix is not supported
//
// Parameters:
//   - c: Application configuration containing database connection details
//
// Returns:
//   - *bun.DB: Database connection if successful
//   - error: Error if connection fails or URI is invalid
func Open(c *config.Config) (db *bun.DB, err error) {
	dsn := c.DB.DatabaseURI

	if dsn == "" {
		fmt.Println("not using database")
		return
	}

	found := false
	var usedPrefix string

	for _, bunFactory := range bunFactories {
		for _, prefix := range bunFactory.Prefixes {
			if found = strings.HasPrefix(dsn, prefix); found {
				usedPrefix = prefix
				break
			}
		}

		if found {
			db, err = bunFactory.Opener(c)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to database with prefix %s: %w", usedPrefix, err)
			}
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("invalid database connection string %s, only (%s) is supported", dsn, allPrefixes())
	}

	return db, nil
}
