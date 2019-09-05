package query

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// Pagination constants
const (
	// ParamPage is a JSON API query parameter used as for pagination.
	ParamPage = "page"
	// ParamPageNumber is a JSON API query parameter used in a page based
	// pagination strategy in conjunction with ParamPageSize.
	ParamPageNumber = "page[number]"
	// ParamPageSize is a JSON API query parameter used in a page based
	// pagination strategy in conjunction with ParamPageNumber.
	ParamPageSize = "page[size]"
	// ParamPageOffset is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with ParamPageLimit.
	ParamPageOffset = "page[offset]"
	// ParamPageLimit is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with ParamPageOffset.
	ParamPageLimit = "page[limit]"
	// ParamPageCursor is a JSON API query parameter used with a cursor-based
	// strategy.
	ParamPageCursor = "page[cursor]"
	// ParamPageTotal is a JSON API query parameter used in pagination
	// It tells to API to add information about total-pages or total-count
	// (depending on the current strategy).
	ParamPageTotal = "page[total]"
)

// PaginationType defines the pagination type.
type PaginationType int

const (
	// LimitOffsetPagination is the pagination type that defines limit or (and) offset.
	LimitOffsetPagination PaginationType = iota
	// PageNumberPagination is the pagination type that uses page type pagination i.e. page=1 page size = 10.
	PageNumberPagination
)

// Pagination defines the query limits and offsets.
// It defines the maximum size (Limit) as well as an offset at which
// the query should start.
type Pagination struct {
	// Size is a pagination value that defines 'limit' or 'page size'
	Size int
	// Offset is a pagination value that defines 'offset' or 'page number'
	Offset int
	Type   PaginationType
}

// Check checks if the pagination is well formed.
func (p *Pagination) Check() error {
	return p.checkValues()
}

// IsZero checks if the pagination is already set.
func (p *Pagination) IsZero() bool {
	return p.Size == 0 && p.Offset == 0
}

// FormatQuery formats the pagination for the url query.
func (p *Pagination) FormatQuery(q ...url.Values) url.Values {
	var query url.Values
	if len(q) != 0 {
		query = q[0]
	}

	if query == nil {
		query = url.Values{}
	}

	var k, v string
	switch p.Type {
	case LimitOffsetPagination:
		limit, offset := p.Size, p.Offset
		if limit != 0 {
			k = ParamPageLimit
			v = strconv.Itoa(limit)
			query.Set(k, v)
		}
		if offset != 0 {
			k = ParamPageOffset
			v = strconv.Itoa(offset)
			query.Set(k, v)
		}
	case PageNumberPagination:
		number, size := p.Offset, p.Size
		if number != 0 {
			k = ParamPageNumber
			v = strconv.Itoa(number)
			query.Set(k, v)
		}

		if size != 0 {
			k = ParamPageSize
			v = strconv.Itoa(size)
			query.Set(k, v)
		}
	default:
		log.Debugf("Pagination with invalid type: '%s'", p.Type)
	}
	return query
}

// String implements fmt.Stringer interface.
func (p *Pagination) String() string {
	switch p.Type {
	case LimitOffsetPagination:
		return fmt.Sprintf("Limit: %d, Offset: %d", p.Size, p.Offset)
	case PageNumberPagination:
		return fmt.Sprintf("PageSize: %d, PageNumber: %d", p.Offset, p.Size)
	default:
		return "Unknown Pagination"
	}
}

func (p *Pagination) checkValues() error {
	switch p.Type {
	case LimitOffsetPagination:
		return p.checkOffsetBasedValues()
	case PageNumberPagination:
		return p.checkPageBasedValues()
	default:
		return errors.NewDetf(class.QueryPaginationType, "unsupported pagination type: '%v'", p.Type)
	}
}

func (p *Pagination) checkOffsetBasedValues() error {
	if p.Size < 0 {
		err := errors.NewDet(class.QueryPaginationValue, "invalid pagination")
		err.SetDetails("Pagination limit lower than -1")
		return err
	}

	if p.Offset < 0 {
		err := errors.NewDet(class.QueryPaginationValue, "invalid pagination")
		err.SetDetails("Pagination offset lower than 0")
		return err
	}
	return nil
}

func (p *Pagination) checkPageBasedValues() error {
	if p.Size < 0 {
		err := errors.NewDet(class.QueryPaginationValue, "invalid pagination value")
		err.SetDetails("Pagination page-size lower than 0")
		return err
	}

	if p.Offset < 0 {
		err := errors.NewDet(class.QueryPaginationValue, "invalid pagination value")
		err.SetDetails("Pagination page-number lower than 0")
		return err
	}
	return nil
}
