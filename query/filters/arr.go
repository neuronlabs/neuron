package filters

import (
	"strings"
)

// Filters is the wrapper over the slice of filter fields.
type Filters []*FilterField

// String implements fmt.Stringer interface.
func (f Filters) String() string {
	sb := &strings.Builder{}

	for _, ff := range f {
		ff.buildString(sb, 0)
	}
	return sb.String()
}
