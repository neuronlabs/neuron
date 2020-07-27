package filter

import (
	"strings"
)

// Or creates a logical 'OR' group filter. Currently only simple filter is allowed to be part of OrGroupFilter
func Or(filters ...Simple) Filter {
	return OrGroup(filters)
}

// OrGroup is the filter that allows to make a logical OR query.
type OrGroup []Simple

// Copy implements Filter interface.
func (o OrGroup) Copy() Filter {
	cp := make([]Simple, len(o))
	for i, filter := range o {
		cp[i] = filter.Copy().(Simple)
	}
	return OrGroup(cp)
}

// Stringer implements fmt.Stringer interface.
func (o OrGroup) String() string {
	sb := strings.Builder{}
	for i, filter := range o {
		sb.WriteString(filter.String())
		if i != len(o)-1 {
			sb.WriteString(" OR ")
		}
	}
	return sb.String()
}
