package constant

type ProductCategory string

const (
	CategoryCoffee    ProductCategory = "COFFEE"
	CategoryNonCoffee ProductCategory = "NON-COFFEE"
)

func (c ProductCategory) String() string {
	return string(c)
}
