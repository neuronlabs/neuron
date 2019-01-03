package sorts

// Order is an enumerator that describes the order of sorting
type Order int

const (
	AscendingOrder Order = iota
	DescendingOrder
)

// String implements Stringer interface
func (o Order) String() string {
	if o == AscendingOrder {
		return "ascending"
	}
	return "descending"
}
