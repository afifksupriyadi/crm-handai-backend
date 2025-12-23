package constant

type VariantPricing struct {
	Size          VariantSize
	PriceModifier float64
	IsDefault     bool
}

var VariantPricingRules = map[string][]VariantPricing{
	"Americano (Botol)": {
		{Size: VariantSize250ml, PriceModifier: 0, IsDefault: true},
		{Size: VariantSize500ml, PriceModifier: 10000, IsDefault: false},
		{Size: VariantSize1000ml, PriceModifier: 27000, IsDefault: false},
	},
	"Choco Latte (Botol)": {
		{Size: VariantSize250ml, PriceModifier: 0, IsDefault: true},
		{Size: VariantSize500ml, PriceModifier: 13000, IsDefault: false},
		{Size: VariantSize1000ml, PriceModifier: 39000, IsDefault: false},
	},
	"Kopi Susu Gula Aren (Botol)": {
		{Size: VariantSize250ml, PriceModifier: 0, IsDefault: true},
		{Size: VariantSize500ml, PriceModifier: 13000, IsDefault: false},
		{Size: VariantSize1000ml, PriceModifier: 39000, IsDefault: false},
	},
	"Matcha Latte (Botol)": {
		{Size: VariantSize250ml, PriceModifier: 0, IsDefault: true},
		{Size: VariantSize500ml, PriceModifier: 13000, IsDefault: false},
		{Size: VariantSize1000ml, PriceModifier: 39000, IsDefault: false},
	},
	"Susu Kurma (Botol)": {
		{Size: VariantSize250ml, PriceModifier: 0, IsDefault: true},
		{Size: VariantSize500ml, PriceModifier: 13000, IsDefault: false},
		{Size: VariantSize1000ml, PriceModifier: 39000, IsDefault: false},
	},
	"Kopi Susu Gula Aren (Strong) (Botol)": {
		{Size: VariantSize250ml, PriceModifier: 0, IsDefault: true},
		{Size: VariantSize500ml, PriceModifier: 13000, IsDefault: false},
		{Size: VariantSize1000ml, PriceModifier: 39000, IsDefault: false},
	},
}
