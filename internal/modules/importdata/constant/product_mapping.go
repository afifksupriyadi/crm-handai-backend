// internal/modules/importdata/constant/product_mapping.go

package constant

type ProductInfo struct {
	Name        string
	Category    ProductCategory
	BasePrice   float64
	HasVariants bool
}

var ProductMapping = map[string]ProductInfo{
	// === NON-BOTTLE PRODUCTS (23 products - No variants) ===
	"Americano": {
		Name:        "Americano",
		Category:    CategoryCoffee,
		BasePrice:   8000,
		HasVariants: false,
	},
	"Butterscotch Latte": {
		Name:        "Butterscotch Latte",
		Category:    CategoryNonCoffee,
		BasePrice:   15000,
		HasVariants: false,
	},
	"Choco Hazelnut": {
		Name:        "Choco Hazelnut",
		Category:    CategoryCoffee,
		BasePrice:   15000,
		HasVariants: false,
	},
	"Choco Latte": {
		Name:        "Choco Latte",
		Category:    CategoryCoffee,
		BasePrice:   12000,
		HasVariants: false,
	},
	"Es": {
		Name:        "Es",
		Category:    CategoryNonCoffee,
		BasePrice:   2000,
		HasVariants: false,
	},
	"Espresso": {
		Name:        "Espresso",
		Category:    CategoryCoffee,
		BasePrice:   5000,
		HasVariants: false,
	},
	"Gula Aren": {
		Name:        "Gula Aren",
		Category:    CategoryNonCoffee,
		BasePrice:   3000,
		HasVariants: false,
	},
	"Gula Singkong": { // ← NEW PRODUCT
		Name:        "Gula Singkong",
		Category:    CategoryNonCoffee,
		BasePrice:   2000,
		HasVariants: false,
	},
	"Hazelnut Latte": {
		Name:        "Hazelnut Latte",
		Category:    CategoryCoffee,
		BasePrice:   15000,
		HasVariants: false,
	},
	"Kopi Coklat": {
		Name:        "Kopi Coklat",
		Category:    CategoryCoffee,
		BasePrice:   15000,
		HasVariants: false,
	},
	"Kopi Matcha (Dirty Matcha)": {
		Name:        "Kopi Matcha (Dirty Matcha)",
		Category:    CategoryCoffee,
		BasePrice:   15000,
		HasVariants: false,
	},
	"Kopi Susu Aja": {
		Name:        "Kopi Susu Aja",
		Category:    CategoryCoffee,
		BasePrice:   12000,
		HasVariants: false,
	},
	"Kopi Susu Gula Aren": {
		Name:        "Kopi Susu Gula Aren",
		Category:    CategoryCoffee,
		BasePrice:   12000,
		HasVariants: false,
	},
	"Kopi Susu Madu": {
		Name:        "Kopi Susu Madu",
		Category:    CategoryCoffee,
		BasePrice:   15000,
		HasVariants: false,
	},
	"Kopi Tubruk": {
		Name:        "Kopi Tubruk",
		Category:    CategoryCoffee,
		BasePrice:   5000,
		HasVariants: false,
	},
	"Madu": {
		Name:        "Madu",
		Category:    CategoryNonCoffee,
		BasePrice:   3000,
		HasVariants: false,
	},
	"Matcha Latte": {
		Name:        "Matcha Latte",
		Category:    CategoryNonCoffee,
		BasePrice:   12000,
		HasVariants: false,
	},
	"Matcha Latte Banget": {
		Name:        "Matcha Latte Banget",
		Category:    CategoryNonCoffee,
		BasePrice:   17000,
		HasVariants: false,
	},
	"Stevia": {
		Name:        "Stevia",
		Category:    CategoryNonCoffee,
		BasePrice:   1000,
		HasVariants: false,
	},
	"Susu": {
		Name:        "Susu",
		Category:    CategoryNonCoffee,
		BasePrice:   3000,
		HasVariants: false,
	},
	"Susu Kurma": {
		Name:        "Susu Kurma",
		Category:    CategoryNonCoffee,
		BasePrice:   12000,
		HasVariants: false,
	},
	"Usucha": {
		Name:        "Usucha",
		Category:    CategoryNonCoffee,
		BasePrice:   10000,
		HasVariants: false,
	},
	"Vanilla Latte": {
		Name:        "Vanilla Latte",
		Category:    CategoryCoffee,
		BasePrice:   15000,
		HasVariants: false,
	},

	// === BOTTLE PRODUCTS (5 products - With variants) ===
	"Americano (Botol)": {
		Name:        "Americano (Botol)",
		Category:    CategoryCoffee,
		BasePrice:   12000,
		HasVariants: true,
	},
	"Choco Latte (Botol)": {
		Name:        "Choco Latte (Botol)",
		Category:    CategoryCoffee,
		BasePrice:   15000,
		HasVariants: true,
	},
	"Kopi Susu Gula Aren (Botol)": {
		Name:        "Kopi Susu Gula Aren (Botol)",
		Category:    CategoryCoffee,
		BasePrice:   15000,
		HasVariants: true,
	},
	"Matcha Latte (Botol)": {
		Name:        "Matcha Latte (Botol)",
		Category:    CategoryNonCoffee,
		BasePrice:   15000,
		HasVariants: true,
	},
	"Susu Kurma (Botol)": {
		Name:        "Susu Kurma (Botol)",
		Category:    CategoryNonCoffee,
		BasePrice:   15000,
		HasVariants: true,
	},
}
