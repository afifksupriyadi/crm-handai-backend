package routes

import (
	"fmt"
	"net/http"

	"github.com/afifksupriyadi/crm-handai-backend/config"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/handler"
	"github.com/danielgtaylor/huma/v2"
)

// RegisterCustomerRoutes registers customer routes with the provided API and handler
func RegisterCustomerRoutes(api huma.API, h *handler.CustomerHandler) {
	basePath := fmt.Sprintf("%s/customers", config.Get().BasePath)

	// POST /api/customers - Create new customer
	huma.Register(api,
		huma.Operation{
			OperationID: "createCustomer",
			Method:      http.MethodPost,
			Path:        basePath,
			Summary:     "Create Customer",
			Description: "Create a new customer with name and phone number.",
			Tags:        []string{"customers"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleCreateCustomer,
	)

	// GET /api/customers - Get all customers with pagination
	huma.Register(api,
		huma.Operation{
			OperationID: "getAllCustomers",
			Method:      http.MethodGet,
			Path:        basePath,
			Summary:     "Get All Customers",
			Description: "Get all customers with pagination and search functionality.",
			Tags:        []string{"customers"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetAllCustomers,
	)

	// GET /api/customers/{id} - Get customer by ID
	huma.Register(api,
		huma.Operation{
			OperationID: "getCustomerByID",
			Method:      http.MethodGet,
			Path:        basePath + "/{id}",
			Summary:     "Get Customer by ID",
			Description: "Retrieve a single customer by their ID.",
			Tags:        []string{"customers"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetCustomerByID,
	)

	// PUT /api/customers/{id} - Update customer
	huma.Register(api,
		huma.Operation{
			OperationID: "updateCustomer",
			Method:      http.MethodPut,
			Path:        basePath + "/{id}",
			Summary:     "Update Customer",
			Description: "Update an existing customer's information.",
			Tags:        []string{"customers"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleUpdateCustomer,
	)

	// DELETE /api/customers/{id} - Delete customer (soft delete)
	huma.Register(api,
		huma.Operation{
			OperationID: "deleteCustomer",
			Method:      http.MethodDelete,
			Path:        basePath + "/{id}",
			Summary:     "Delete Customer",
			Description: "Soft delete a customer by ID.",
			Tags:        []string{"customers"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleDeleteCustomer,
	)

	// NEW: GET /api/customers/recent-transactions - Get customers with recent transactions
	huma.Register(api,
		huma.Operation{
			OperationID: "getCustomersWithRecentTransactions",
			Method:      http.MethodGet,
			Path:        basePath + "/recent-transactions",
			Summary:     "Get Customers With Recent Transactions",
			Description: "Get all customers who have made transactions, ordered by most recent transaction date.",
			Tags:        []string{"customers"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
		}, h.HandleGetCustomersWithRecentTransactions,
	)
}
