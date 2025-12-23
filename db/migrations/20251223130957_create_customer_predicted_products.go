package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Create customer_predicted_products table
		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS analytics.customer_predicted_products (
				id SERIAL PRIMARY KEY,
				prediction_id INT NOT NULL,
				product_id INT NOT NULL,
				variant_id INT NULL,
				predicted_quantity INT NOT NULL DEFAULT 1,
				confidence_score NUMERIC(5,2) NOT NULL DEFAULT 0.00,
				purchase_frequency INT NOT NULL DEFAULT 0,
				avg_quantity NUMERIC(5,2) NOT NULL DEFAULT 1.00,
				
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				
				CONSTRAINT fk_predicted_product_prediction 
					FOREIGN KEY (prediction_id) 
					REFERENCES analytics.customer_predictions(id) 
					ON DELETE CASCADE,
				CONSTRAINT fk_predicted_product_product 
					FOREIGN KEY (product_id) 
					REFERENCES products(id)
					ON DELETE RESTRICT,
				CONSTRAINT fk_predicted_product_variant 
					FOREIGN KEY (variant_id) 
					REFERENCES variants(id)
					ON DELETE RESTRICT,
				CONSTRAINT chk_predicted_quantity 
					CHECK (predicted_quantity > 0),
				CONSTRAINT chk_confidence_score 
					CHECK (confidence_score >= 0 AND confidence_score <= 100)
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create customer_predicted_products table: %w", err)
		}

		// Create indexes
		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_predicted_products_prediction 
			ON analytics.customer_predicted_products(prediction_id)
		`)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}

		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_predicted_products_product 
			ON analytics.customer_predicted_products(product_id)
		`)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// Rollback: Drop table (CASCADE will handle predicted products)
		_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS analytics.customer_predicted_products CASCADE`)
		return err
	})
}
