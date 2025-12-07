package constant

type PackageType string

const (
	PackageTypeCup    PackageType = "CUP"
	PackageTypeBottle PackageType = "BOTTLE"
)

func (p PackageType) String() string {
	return string(p)
}
