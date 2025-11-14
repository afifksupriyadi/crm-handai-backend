package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/config"
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
			_, err = db.NewSelect().ColumnExpr("1").Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("error connecting to postgresql: %w", err)
			}

			slog.Info("postgresql connected successfully")

			return db, nil
		},
	})
}
