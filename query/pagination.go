package query

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/neuronlabs/neuron/errors"
)

// Pagination defined constants used for formatting the query.
const (
	// ParamPage is a JSON API query parameter used as for pagination.
	ParamPage = "page"
	// ParamPageOffset is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with ParamPageLimit.
	ParamPageOffset = "page[offset]"
	// ParamPageLimit is a JSON API query parameter used in an offset based
	// pagination strategy in conjunction with ParamPageOffset.
	ParamPageLimit = "page[limit]"
)

// Limit sets the maximum number of objects returned by the Find process,
// Returns error if the given scope has already different type of pagination.
func (s *Scope) Limit(limit int64) {
	if s.Pagination == nil {
		s.Pagination = &Pagination{}
	}
	s.Pagination.Limit = limit
}

// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as omitting the 'Offset' clause.
// Returns error if the given scope has already different type of pagination.
func (s *Scope) Offset(offset int64) {
	if s.Pagination == nil {
		s.Pagination = &Pagination{}
	}
	s.Pagination.Offset = offset
}

// Pagination defines the query limits and offsets.
// It defines the maximum size (Limit) as well as an offset at which
// the query should start.
// If the pagination type is 'LimitOffsetPagination' the value of 'Limit' defines the 'limit'
// where the value of 'Offset' defines it's 'offset'.
// If the pagination type is 'PageNumberPagination' the value of 'Limit' defines 'pageSize'
// and the value of 'Offset' defines 'pageNumber'. The page number value starts from '1'.
type Pagination struct {
	// Limit is a pagination value that defines 'limit' or 'page size'
	Limit int64
	// Offset is a pagination value that defines 'offset' or 'page number'
	Offset int64
}

// First gets the first pagination for provided 'p' pagination values.
// If the 'p' pagination is already the 'first' pagination the function
// returns it as the result directly.
func (p *Pagination) First() (*Pagination, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if p.Offset == 0 {
		return p, nil
	}
	return &Pagination{Limit: p.Limit, Offset: 0}, nil
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
	limit, offset := p.Limit, p.Offset
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
	return query
}

// Validate checks if the pagination is well formed.
func (p *Pagination) Validate() error {
	return p.checkValues()
}

// IsZero checks if the pagination is zero valued.
func (p *Pagination) IsZero() bool {
	return p.Limit == 0 && p.Offset == 0
}

// Last gets the last pagination for the provided 'total' count.
// Returns error if current pagination is not valid, or 'total'.
// If current pagination 'p' is the last pagination it would be return
// directly as the result. In order to check if the 'p' is last pagination
// compare it's pointer with the 'p'.
func (p *Pagination) Last(total int64) (*Pagination, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if total < 0 {
		return nil, errors.NewDetf(ClassInvalidInput, "Total instance value lower than 0: %v", total)
	}

	var offset int64
	// check if the last page is not partial
	if partialSize := p.Offset % p.Limit; partialSize != 0 {
		lastFull := (total - partialSize) / p.Limit
		offset = lastFull*p.Limit + partialSize
	} else {
		// the last should be total/p.Limit - (10-3)/2 = 3
		// 3 * size = 3 * 2 = 6
		offset = total - p.Limit
		// in case when total size is lower then the pagination size set the offset to 0
		if offset < 0 {
			offset = 0
		}
	}
	if offset == p.Offset {
		return p, nil
	}
	// the offset should be total / p.Limit
	return &Pagination{Limit: p.Limit, Offset: offset}, nil
}

// Next gets the next pagination for the provided 'total' count.
// If current pagination 'p' is the last one, then the function returns 'p' pagination directly.
// In order to check if there is a next pagination compare the result pointer with the 'p' pagination.
func (p *Pagination) Next(total int64) (*Pagination, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if total < 0 {
		return nil, errors.NewDetf(ClassInvalidInput, "Total instance value lower than 0: %v", total)
	}

	// keep the same size but change the offset
	offset := p.Offset + p.Limit
	// check if the next offset doesn't overflow 'total'
	// in example:
	// 	total 52; p.Offset = 50; p.Limit = 10
	//	the next offset would be 60 which overflows possible total values.
	if offset >= total {
		return p, nil
	}
	return &Pagination{Offset: offset, Limit: p.Limit}, nil
}

// Previous gets the pagination for the previous possible size and offset.
// If current pagination 'p' is the first page then the function returns 'p' pagination.
// If the previous size would overflow the 0th offset then the previous starts from 0th offset.
func (p *Pagination) Previous() (*Pagination, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	if p.Offset == 0 {
		return p, nil
	}
	if p.Offset < p.Limit {
		return &Pagination{Offset: 0, Limit: p.Offset}, nil
	}
	// keep the same size but change the offset
	return &Pagination{Offset: p.Offset - p.Limit, Limit: p.Limit}, nil
}

// String implements fmt.Stringer interface.
func (p *Pagination) String() string {
	return fmt.Sprintf("Limit: %d, Offset: %d", p.Limit, p.Offset)
}

func (p *Pagination) checkValues() error {
	return p.checkOffsetBasedValues()
}

func (p *Pagination) checkOffsetBasedValues() error {
	if p.Limit < 0 {
		err := errors.NewDet(ClassInvalidInput, "invalid pagination")
		err.SetDetails("Pagination limit lower than -1")
		return err
	}

	if p.Offset < 0 {
		err := errors.NewDet(ClassInvalidInput, "invalid pagination")
		err.SetDetails("Pagination offset lower than 0")
		return err
	}
	return nil
}
