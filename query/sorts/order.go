package sorts

// Order is the enum used for the sorting values
type Order int

const (
	// AscendingOrder defines the sorting ascending order
	AscendingOrder Order = iota

	// DescendingOrder defines the sorting descending order
	DescendingOrder
)

// String implements Stringer interface
func (o Order) String() string {
	if o == AscendingOrder {
		return "ascending"
	}

	return "descending"
}
