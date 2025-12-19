package constant

// CustomerSegment represents customer segmentation categories
type CustomerSegment string

const (
	SegmentNew       CustomerSegment = "NEW"
	SegmentPotential CustomerSegment = "POTENTIAL"
	SegmentLoyal     CustomerSegment = "LOYAL"
	SegmentChurn     CustomerSegment = "CHURN"
)

// String returns string representation
func (s CustomerSegment) String() string {
	return string(s)
}
