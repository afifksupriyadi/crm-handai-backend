// internal/modules/customer/repository/customer_repository.go

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/uptrace/bun"
)

type CustomerRepository interface {
	GetCustomerByID(ctx context.Context, id int) (*model.Customer, error)
	GetCustomerByPhone(ctx context.Context, phone string) (*model.Customer, error)
	GetCustomerByName(ctx context.Context, name string) (*model.Customer, error)
	CreateCustomer(ctx context.Context, customer *model.Customer) error
	GetCustomerCount(ctx context.Context) (int, error)
}

type CustomerRepositoryImpl struct {
	db *bun.DB
}

func NewCustomerRepository(db *bun.DB) CustomerRepository {
	return &CustomerRepositoryImpl{db: db}
}

func (r *CustomerRepositoryImpl) GetCustomerByID(ctx context.Context, id int) (*model.Customer, error) {
	customer := new(model.Customer)
	err := r.db.NewSelect().
		Model(customer).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

func (r *CustomerRepositoryImpl) GetCustomerByPhone(ctx context.Context, phone string) (*model.Customer, error) {
	customer := new(model.Customer)
	err := r.db.NewSelect().
		Model(customer).
		Where("phone = ?", phone).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

func (r *CustomerRepositoryImpl) GetCustomerByName(ctx context.Context, name string) (*model.Customer, error) {
	customer := new(model.Customer)
	err := r.db.NewSelect().
		Model(customer).
		Where("name = ?", name).
		Where("deleted_at IS NULL").
		Scan(ctx)

	return customer, err
}

func (r *CustomerRepositoryImpl) CreateCustomer(ctx context.Context, customer *model.Customer) error {
	_, err := r.db.NewInsert().
		Model(customer).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *CustomerRepositoryImpl) GetCustomerCount(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().
		Model((*model.Customer)(nil)).
		Where("deleted_at IS NULL").
		Count(ctx)

	return count, err
}
