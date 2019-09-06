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
// If the pagination type is 'LimitOffsetPagination' the value of 'Size' defines the 'limit'
// where the value of 'Offset' defines it's 'offset'.
// If the pagination type is 'PageNumberPagination' the value of 'Size' defines 'pageSize'
// and the value of 'Offset' defines 'pageNumber'. The page number value starts from '1'.
type Pagination struct {
	// Size is a pagination value that defines 'limit' or 'page size'
	Size int
	// Offset is a pagination value that defines 'offset' or 'page number'
	Offset int
	Type   PaginationType
}

// IsValid checks if the pagination is well formed.
func (p *Pagination) IsValid() error {
	return p.checkValues()
}

// GetLimitOffset gets the 'limit' and 'offset' values from the given pagination.
// If the pagination type is 'LimitOffsetPagination' then the value of 'limit' = p.Size
// and the value of 'offset' = p.Offset.
// In case when pagination type is 'PageNumberPagination' the 'limit' = p.Size and the
// offset is a result of multiplication of (pageNumber - 1) * pageSize = (p.Offset - 1) * p.Size.
// If the p.Offset is zero value then the (pageNumber - 1) value would be set previously to '0'.
func (p *Pagination) GetLimitOffset() (limit, offset int) {
	switch p.Type {
	case LimitOffsetPagination:
		limit = p.Size
		offset = p.Offset
	case PageNumberPagination:
		limit = p.Size
		// the p.Offset value is a page number that starts it's value from '1'.
		// If it's value is zero - the offset = 0.
		if p.Offset >= 1 {
			offset = (p.Offset - 1) * p.Size
		}
	default:
		log.Warningf("Unknown pagination type: '%v'", p.Type)
	}
	return limit, offset
}

// GetNumberSize gets the 'page number' and 'page size' from the pagination.
// If the pagination type is 'PageNumberPagination' the results are just the values of pagination.
// In case the pagination type is of 'LimitOffsetPagination' then 'limit'(Size) would be the page size
// and the page number would be 'offset' / limit + 1. PageNumberPagination starts it's page number counting from 1.
// If the offset % size != 0 - the offset is not dividable by the size without the rest - then the division
// rounds down it's value.
func (p *Pagination) GetNumberSize() (pageNumber, pageSize int) {
	switch p.Type {
	case LimitOffsetPagination:
		pageSize = p.Size
		// the default pageNumber value is '1'.
		pageNumber = 1
		// if the 'limit' and 'offset' values are greater than zero - compute the value of pageNumber.
		if p.Size > 0 && p.Offset > 0 {
			// page numbering starts from '1' thus the result would be
			pageNumber = p.Offset/p.Size + 1
		}
	case PageNumberPagination:
		pageSize = p.Size
		pageNumber = p.Offset
	default:
		log.Warningf("Unknown pagination type: '%v'", p.Type)
	}
	return pageNumber, pageSize
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
		if number >= 1 {
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

	if p.Offset <= 0 {
		err := errors.NewDet(class.QueryPaginationValue, "invalid pagination value")
		err.SetDetails("Pagination page-number lower equalt to 0")
		return err
	}
	return nil
}
