package query

import (
	"github.com/kucjac/jsonapi/pkg/internal/query/paginations"
)

// Pagination defines the query limits and offsets.
// It defines the maximum size (Limit) as well as an offset at which
// the query should start.
type Pagination struct {
	*paginations.Pagination
}
