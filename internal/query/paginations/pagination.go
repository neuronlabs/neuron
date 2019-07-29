package paginations

import (
	"fmt"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// Type is the pagination type definition enum.
type Type int

const (
	// TpOffset is the 'offset' based pagination type.
	TpOffset Type = iota
	// TpPage is the 'page' based pagination type.
	TpPage
)

// String implements fmt.Stringer interface.
func (t Type) String() string {
	switch t {
	case TpOffset:
		return "Offset"
	case TpPage:
		return "Page"
	default:
		return "Unknown"
	}
}

// Parameter Defines the given paginate paramater type.
type Parameter int

const (
	// ParamLimit defines 'limit' of the pagination. Used for 'offset' based paginations.
	ParamLimit Parameter = iota

	// ParamOffset defines 'offset' of the pagination. Used for 'offset' based paginations.
	ParamOffset

	// ParamNumber defines the 'page number' of the pagination. Used for 'page' based paginations.
	ParamNumber

	// ParamSize defines 'page size' of the pagination. Used for 'page' based paginations.
	ParamSize
)

// Pagination is a struct that keeps variables used to paginate the result.
type Pagination struct {
	first  int // first is the limit, or page number
	second int // second is the offset or page size

	// Describes which pagination type to use.
	tp Type
}

// NewLimitOffset creates new limit offset pagination.
func NewLimitOffset(limit, offset int) *Pagination {
	return &Pagination{
		first:  limit,
		second: offset,
		tp:     TpOffset,
	}
}

// NewPaged creates new page based Pagination.
func NewPaged(pageNumber, pageSize int) *Pagination {
	return &Pagination{
		first:  pageNumber,
		second: pageSize,
		tp:     TpPage,
	}
}

// Check checks if the given pagination is valid.
func (p *Pagination) Check() error {
	return p.check()
}

// IsZero checks if the pagination were already set.
func (p *Pagination) IsZero() bool {
	return p.first == 0 && p.second == 0 && p.tp == 0
}

// SetType sets the pagination type.
func (p *Pagination) SetType(tp Type) {
	p.tp = tp
}

// SetLimitOffset sets the Pagination from the PageType into OffsetType.
func (p *Pagination) SetLimitOffset() {
	if p.tp == TpPage {
		p.first = p.second
		p.second = p.first * p.second
		p.tp = TpOffset
	}
}

// SetValue sets the pagination value for given parameter.
func (p *Pagination) SetValue(val int, parameter Parameter) {
	switch parameter {
	case ParamLimit, ParamNumber:
		p.first = val
	case ParamOffset, ParamSize:
		p.second = val
	}
}

// Type gets the pagination type.
func (p *Pagination) Type() Type {
	return p.tp
}

// String implements fmt.Stringer interface.
func (p *Pagination) String() string {
	switch p.tp {
	case TpOffset:
		return fmt.Sprintf("Offset Pagination. Limit: %d, Offset: %d", p.first, p.second)
	case TpPage:
		return fmt.Sprintf("PageType Pagination. PageSize: %d, PageNumber: %d", p.second, p.first)
	default:
		return "Unknown Pagination"
	}
}

// GetLimitOffset gets the 'limit' and 'offset' values from the given pagination.
func (p *Pagination) GetLimitOffset() (limit, offset int) {
	switch p.tp {
	case TpOffset:
		limit = p.first
		offset = p.second
	case TpPage:
		limit = p.second
		offset = p.second * p.first
	default:
		log.Warningf("Unknown pagination type: '%v'", p.tp)
	}
	return limit, offset
}

// GetNumberSize gets the 'page number' and 'page size' from the pagination.
func (p *Pagination) GetNumberSize() (pageNumber, pageSize int) {
	switch p.tp {
	case TpOffset:
		pageSize = p.second
		pageNumber = p.second / p.first
	case TpPage:
		pageNumber = p.first
		pageSize = p.second
	}
	return
}

func (p *Pagination) check() error {
	if err := p.checkValues(); err != nil {
		return err
	}
	return nil
}

func (p *Pagination) checkValues() error {
	switch p.Type() {
	case TpOffset:
		return p.checkOffsetBasedValues()
	case TpPage:
		return p.checkPageBasedValues()
	default:
		return errors.NewDetf(class.QueryPaginationType, "unsupported pagination type: '%v'", p.Type())
	}
}

func (p *Pagination) checkOffsetBasedValues() error {
	if p.first < 0 {
		err := errors.NewDet(class.QueryPaginationValue, "invalid pagination")
		err.SetDetails("Pagination limit lower than -1")
		return err
	}

	if p.second < 0 {
		err := errors.NewDet(class.QueryPaginationValue, "invalid pagination")
		err.SetDetails("Pagination offset lower than 0")
		return err
	}
	return nil
}

func (p *Pagination) checkPageBasedValues() error {
	if p.first < 0 {
		err := errors.NewDet(class.QueryPaginationValue, "invalid pagination value")
		err.SetDetails("Pagination page-number lower than 0")
		return err
	}

	if p.second < 0 {
		err := errors.NewDet(class.QueryPaginationValue, "invalid pagination value")
		err.SetDetails("Pagination page-size lower than 0")
		return err
	}
	return nil
}
