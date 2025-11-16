package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func init() {
	bunFactories = append(bunFactories, &BunFactory{
		Prefixes: []string{
			"postgres://", "postgresql://", "unix://",
		},
		Opener: func(c *config.Config) (db *bun.DB, err error) {
			dsn := c.DB.DatabaseURI

			var dbConn *sql.DB
			connector := pgdriver.NewConnector(
				pgdriver.WithDSN(dsn),
				pgdriver.WithTimeout(time.Duration(c.DB.DatabaseTimeout)*time.Second),
			)
			dbConn = sql.OpenDB(connector)

			// Configure connection pool
			dbConn.SetMaxOpenConns(c.DB.DatabaseMaxOpenConns)
			dbConn.SetMaxIdleConns(c.DB.DatabaseMaxIdleConns)
			dbConn.SetConnMaxLifetime(c.DB.DatabaseConnMaxLifetime)

			db = bun.NewDB(dbConn, pgdialect.New(), bun.WithDiscardUnknownColumns())

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.DB.DatabaseTimeout)*time.Second)
			defer cancel()

			// health check connection
			logger.Get().Debug().Msg("Executing database health check query")

			_, err = db.NewSelect().ColumnExpr("1").Exec(ctx)
			if err != nil {
				logger.Get().Error().
					Err(err).
					Int("timeout_seconds", c.DB.DatabaseTimeout).
					Msg("Database health check failed")
				return nil, fmt.Errorf("error connecting to postgresql: %w", err)
			}

			logger.Get().Debug().Msg("Database health check passed")

			return db, nil
		},
	})
}
