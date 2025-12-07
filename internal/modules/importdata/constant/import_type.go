package constant

type ImportType string

const (
	ImportTypeCustomer    ImportType = "CUSTOMER"
	ImportTypeTransaction ImportType = "TRANSACTION"
)

func (t ImportType) String() string {
	return string(t)
}
