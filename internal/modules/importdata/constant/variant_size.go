package constant

type VariantSize string

const (
	VariantSize250ml  VariantSize = "250ml"
	VariantSize500ml  VariantSize = "500ml"
	VariantSize1000ml VariantSize = "1000ml"
)

func (v VariantSize) String() string {
	return string(v)
}
