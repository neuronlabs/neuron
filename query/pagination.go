package query

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// Pagination defined constants used for formatting the query.
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
)

// PaginationType defines the pagination type.
type PaginationType int

const (
	// LimitOffsetPagination is the pagination type that defines limit or (and) offset.
	LimitOffsetPagination PaginationType = iota
	// PageNumberPagination is the pagination type that uses page type pagination i.e. page=1 page size = 10.
	PageNumberPagination
)

// NewPaginationLimitOffset creates new Limit Offset pagination for given 'limit' and 'offset'.
func NewPaginationLimitOffset(limit, offset int64) (*Pagination, error) {
	pagination := &Pagination{Type: LimitOffsetPagination, Size: limit, Offset: offset}
	if err := pagination.IsValid(); err != nil {
		return nil, err
	}
	return pagination, nil
}

// NewPaginationNumberSize sets the pagination of the type PageNumberSize with the page 'number' and page 'size'.
func NewPaginationNumberSize(number, size int64) (*Pagination, error) {
	pagination := &Pagination{Size: size, Offset: number, Type: PageNumberPagination}
	if err := pagination.IsValid(); err != nil {
		return nil, err
	}
	return pagination, nil
}

// Pagination defines the query limits and offsets.
// It defines the maximum size (Limit) as well as an offset at which
// the query should start.
// If the pagination type is 'LimitOffsetPagination' the value of 'Size' defines the 'limit'
// where the value of 'Offset' defines it's 'offset'.
// If the pagination type is 'PageNumberPagination' the value of 'Size' defines 'pageSize'
// and the value of 'Offset' defines 'pageNumber'. The page number value starts from '1'.
type Pagination struct {
	// Size is a pagination value that defines 'limit' or 'page size'
	Size int64
	// Offset is a pagination value that defines 'offset' or 'page number'
	Offset int64
	Type   PaginationType
}

// First gets the first pagination for provided 'p' pagination values.
// If the 'p' pagination is already the 'first' pagination the function
// returns it as the result directly.
func (p *Pagination) First() (*Pagination, error) {
	if err := p.IsValid(); err != nil {
		return nil, err
	}
	var first *Pagination
	switch p.Type {
	case LimitOffsetPagination:
		if p.Offset == 0 {
			return p, nil
		}
		first = &Pagination{Size: p.Size, Offset: 0, Type: LimitOffsetPagination}
	case PageNumberPagination:
		if p.Offset == 1 {
			return p, nil
		}
		first = &Pagination{Size: p.Size, Offset: 1, Type: PageNumberPagination}
	default:
		return nil, errors.NewDet(class.QueryPaginationType, "invalid pagination type")
	}
	return first, nil
}

// FormatQuery formats the pagination for the url query with respect to JSONAPI specification.
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
			v = strconv.FormatInt(limit, 10)
			query.Set(k, v)
		}
		if offset != 0 {
			k = ParamPageOffset
			v = strconv.FormatInt(offset, 10)
			query.Set(k, v)
		}
	case PageNumberPagination:
		number, size := p.Offset, p.Size
		if number >= 1 {
			k = ParamPageNumber
			v = strconv.FormatInt(number, 10)
			query.Set(k, v)
		}

		if size != 0 {
			k = ParamPageSize
			v = strconv.FormatInt(size, 10)
			query.Set(k, v)
		}
	default:
		log.Debugf("Pagination with invalid type: '%s'", p.Type)
	}
	return query
}

// GetLimitOffset gets the 'limit' and 'offset' values from the given pagination.
// If the pagination type is 'LimitOffsetPagination' then the value of 'limit' = p.Size
// and the value of 'offset' = p.Offset.
// In case when pagination type is 'PageNumberPagination' the 'limit' = p.Size and the
// offset is a result of multiplication of (pageNumber - 1) * pageSize = (p.Offset - 1) * p.Size.
// If the p.Offset is zero value then the (pageNumber - 1) value would be set previously to '0'.
func (p *Pagination) GetLimitOffset() (limit, offset int64) {
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
func (p *Pagination) GetNumberSize() (pageNumber, pageSize int64) {
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

// IsValid checks if the pagination is well formed.
func (p *Pagination) IsValid() error {
	return p.checkValues()
}

// IsZero checks if the pagination is zero valued.
func (p *Pagination) IsZero() bool {
	return p.Size == 0 && p.Offset == 0
}

// Last gets the last pagination for the provided 'total' count.
// Returns error if current pagination is not valid, or 'total'.
// If current pagination 'p' is the last pagination it would be return
// directly as the result. In order to check if the 'p' is last pagination
// compare it's pointer with the 'p'.
func (p *Pagination) Last(total int64) (*Pagination, error) {
	if err := p.IsValid(); err != nil {
		return nil, err
	}
	if total < 0 {
		return nil, errors.NewDetf(class.QueryPaginationValue, "Total instance value lower than 0: %v", total)
	}

	var last *Pagination
	switch p.Type {
	case LimitOffsetPagination:
		var offset int64

		// check if the last page is not partial
		if partialSize := p.Offset % p.Size; partialSize != 0 {
			lastFull := (total - partialSize) / p.Size
			offset = lastFull*p.Size + partialSize
		} else {
			// the last should be total/p.Size - (10-3)/2 = 3
			// 3 * size = 3 * 2 = 6
			offset = total - p.Size
			// in case when total size is lower then the pagination size set the offset to 0
			if offset < 0 {
				offset = 0
			}
		}
		if offset == p.Offset {
			return p, nil
		}
		// the offset should be total / p.Size
		last = &Pagination{Size: p.Size, Offset: offset, Type: LimitOffsetPagination}
	case PageNumberPagination:
		// divide total number of instances by the page size.
		// example:
		// 	total - 52
		//	pageSize - 10
		// 	computedPageNumber = 52/10 = 5
		pageNumber := total / p.Size
		if total%p.Size != 0 || pageNumber == 0 {
			// total % p.Size = 52 % 10 = 2
			pageNumber++
		}
		last = &Pagination{Size: p.Size, Offset: pageNumber, Type: PageNumberPagination}
	default:
		return nil, errors.NewDet(class.QueryPaginationType, "invalid pagination type")
	}
	return last, nil
}

// Next gets the next pagination for the provided 'total' count.
// If current pagination 'p' is the last one, then the function returns 'p' pagination directly.
// In order to check if there is a next pagination compare the result pointer with the 'p' pagination.
func (p *Pagination) Next(total int64) (*Pagination, error) {
	if err := p.IsValid(); err != nil {
		return nil, err
	}
	if total < 0 {
		return nil, errors.NewDetf(class.QueryPaginationValue, "Total instance value lower than 0: %v", total)
	}

	var next *Pagination
	switch p.Type {
	case LimitOffsetPagination:
		// keep the same size but change the offset
		offset := p.Offset + p.Size
		// check if the next offset doesn't overflow 'total'
		// in example:
		// 	total 52; p.Offset = 50; p.Size = 10
		//	the next offset would be 60 which overflows possible total values.
		if offset >= total {
			return p, nil
		}
		next = &Pagination{Offset: offset, Size: p.Size, Type: LimitOffsetPagination}
	case PageNumberPagination:
		// check if the multiplication of pageSize and pageNumber for the current page
		// overflows the 'total'.
		// in example:
		//	total: 52; pageNumber: 6; pageSize: 10;
		//  nextTotal = pageNumber * pageSize = 60
		// 	50 - (10 * 4+1) <= 0 ?
		//
		nextPageNumber := p.Offset + 1
		if total-(p.Size*(nextPageNumber)) <= 0 {
			return p, nil
		}
		next = &Pagination{Offset: nextPageNumber, Size: p.Size, Type: PageNumberPagination}
	default:
		return nil, errors.NewDet(class.QueryPaginationType, "invalid pagination type")
	}
	return next, nil
}

// Previous gets the pagination for the previous possible size and offset.
// If current pagination 'p' is the first page then the function returns 'p' pagination.
// If the previouse size would overflow the 0th offset then the previous starts from 0th offset.
func (p *Pagination) Previous() (*Pagination, error) {
	if err := p.IsValid(); err != nil {
		return nil, err
	}

	var prev *Pagination
	switch p.Type {
	case LimitOffsetPagination:
		if p.Offset == 0 {
			return p, nil
		}
		if p.Offset < p.Size {
			prev = &Pagination{Offset: 0, Size: p.Offset, Type: LimitOffsetPagination}
			return prev, nil
		}
		// keep the same size but change the offset
		prev = &Pagination{Offset: p.Offset - p.Size, Size: p.Size, Type: LimitOffsetPagination}
	case PageNumberPagination:
		if p.Offset <= 1 {
			return p, nil
		}
		prev = &Pagination{Offset: p.Offset - 1, Size: p.Size, Type: PageNumberPagination}
	default:
		return nil, errors.NewDet(class.QueryPaginationType, "invalid pagination type")
	}
	return prev, nil
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
	if p.Size <= 0 {
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
