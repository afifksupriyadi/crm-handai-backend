package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/uptrace/bun"
)

type BatchRepository interface {
	CreateBatch(ctx context.Context, tx *bun.Tx, batch *model.Batch) error
	UpdateBatch(ctx context.Context, tx *bun.Tx, batch *model.Batch) error
	UpdateBatchStatus(ctx context.Context, tx *bun.Tx, batchID int, status string) error
	LinkImportLogs(ctx context.Context, tx *bun.Tx, batchID int, customerImportID, transactionImportID *int) error
	SetActiveBatch(ctx context.Context, tx *bun.Tx, batchID int) error
	GetBatchByID(ctx context.Context, batchID int) (*model.Batch, error)
	GetActiveBatch(ctx context.Context) (*model.Batch, error)
	GetBatchByDate(ctx context.Context, batchDate time.Time) (*model.Batch, error)
	DeleteBatchAndRelatedData(ctx context.Context, tx *bun.Tx, batchID int) error
}

type BatchRepositoryImpl struct {
	db *bun.DB
}

func NewBatchRepository(db *bun.DB) BatchRepository {
	return &BatchRepositoryImpl{db: db}
}

func (r *BatchRepositoryImpl) CreateBatch(ctx context.Context, tx *bun.Tx, batch *model.Batch) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	_, err := db.NewInsert().
		Model(batch).
		Exec(ctx)

	return err
}

func (r *BatchRepositoryImpl) UpdateBatch(ctx context.Context, tx *bun.Tx, batch *model.Batch) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	now := time.Now()
	batch.UpdatedAt = &now

	_, err := db.NewUpdate().
		Model(batch).
		WherePK().
		Exec(ctx)

	return err
}

func (r *BatchRepositoryImpl) UpdateBatchStatus(ctx context.Context, tx *bun.Tx, batchID int, status string) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	_, err := db.NewUpdate().
		Model((*model.Batch)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", batchID).
		Exec(ctx)

	return err
}

func (r *BatchRepositoryImpl) LinkImportLogs(ctx context.Context, tx *bun.Tx, batchID int, customerImportID, transactionImportID *int) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	_, err := db.NewUpdate().
		Model((*model.Batch)(nil)).
		Set("customer_import_id = ?", customerImportID).
		Set("transaction_import_id = ?", transactionImportID).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", batchID).
		Exec(ctx)

	return err
}

func (r *BatchRepositoryImpl) SetActiveBatch(ctx context.Context, tx *bun.Tx, batchID int) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	// First, set all batches to inactive
	_, err := db.NewUpdate().
		Model((*model.Batch)(nil)).
		Set("is_active = ?", false).
		Where("is_active = ?", true).
		Exec(ctx)

	if err != nil {
		return err
	}

	// Then, set the specified batch to active
	_, err = db.NewUpdate().
		Model((*model.Batch)(nil)).
		Set("is_active = ?", true).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", batchID).
		Exec(ctx)

	return err
}

func (r *BatchRepositoryImpl) GetBatchByID(ctx context.Context, batchID int) (*model.Batch, error) {
	batch := new(model.Batch)
	err := r.db.NewSelect().
		Model(batch).
		Where("id = ?", batchID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return batch, err
}

func (r *BatchRepositoryImpl) GetActiveBatch(ctx context.Context) (*model.Batch, error) {
	batch := new(model.Batch)
	err := r.db.NewSelect().
		Model(batch).
		Where("is_active = ?", true).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return batch, err
}

func (r *BatchRepositoryImpl) GetBatchByDate(ctx context.Context, batchDate time.Time) (*model.Batch, error) {
	batch := new(model.Batch)
	err := r.db.NewSelect().
		Model(batch).
		Where("batch_date = ?", batchDate).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return batch, err
}

func (r *BatchRepositoryImpl) DeleteBatchAndRelatedData(ctx context.Context, tx *bun.Tx, batchID int) error {
	var db bun.IDB = r.db
	if tx != nil {
		db = *tx
	}

	// Get batch to find related import logs
	batch := new(model.Batch)
	err := db.NewSelect().
		Model(batch).
		Where("id = ?", batchID).
		Scan(ctx)
	if err != nil {
		return err
	}

	// Delete transaction_details for transactions in this batch
	_, err = db.NewDelete().
		Table("transaction_details").
		Where("transaction_id IN (SELECT id FROM transactions WHERE batch_id = ?)", batchID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Delete transactions in this batch
	_, err = db.NewDelete().
		Table("transactions").
		Where("batch_id = ?", batchID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Delete import logs
	if batch.CustomerImportID != nil {
		_, err = db.NewDelete().
			Table("import_logs").
			Where("id = ?", *batch.CustomerImportID).
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	if batch.TransactionImportID != nil {
		_, err = db.NewDelete().
			Table("import_logs").
			Where("id = ?", *batch.TransactionImportID).
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	// Finally delete the batch itself
	_, err = db.NewDelete().
		Model((*model.Batch)(nil)).
		Where("id = ?", batchID).
		Exec(ctx)

	return err
}
