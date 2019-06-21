package sorts

// Order is an enumerator that describes the order of sorting.
type Order int

const (
	// AscendingOrder is the enum defines ascending sorting order.
	AscendingOrder Order = iota

	// DescendingOrder is the enum that defines descending sorting order.
	DescendingOrder
)

// String implements Stringer interface.
func (o Order) String() string {
	if o == AscendingOrder {
		return "ascending"
	}
	return "descending"
}
