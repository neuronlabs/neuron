package paginations

import (
	"errors"
	"fmt"
)

// Type is the pagination type definition enum
type Type int

// Enum defined for the scope's pagination
const (
	TpOffset Type = iota
	TpPage
)

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

// Parameter Defines the given paginate paramater type
type Parameter int

// Parameter enum definition
const (
	ParamLimit Parameter = iota
	ParamOffset
	ParamNumber
	ParamSize
)

// Pagination is a struct that keeps variables used to paginate the result
type Pagination struct {
	first  int // first is the limit, or page number
	second int // second is the offset or page size

	// Describes which pagination type to use.
	tp Type
}

// NewLimitOffset creates new limit offset pagination
func NewLimitOffset(limit, offset int) *Pagination {
	return &Pagination{
		first:  limit,
		second: offset,
		tp:     TpOffset,
	}
}

// NewPaged creates new pagination of TpPage type
func NewPaged(pageNumber, pageSize int) *Pagination {
	return &Pagination{
		first:  pageNumber,
		second: pageSize,
		tp:     TpPage,
	}
}

// CheckPagination checks if the given pagination is valid
func CheckPagination(p *Pagination) error {
	return p.check()
}

// Check checks if the given pagination is valid
func (p *Pagination) Check() error {
	return p.check()
}

// IsZero checks if the pagination were already set
func (p *Pagination) IsZero() bool {
	return p.first == 0 && p.second == 0 && p.tp == 0
}

// SetType sets the pagination type
func (p *Pagination) SetType(tp Type) {
	p.tp = tp
}

// SetLimitOffset sets the Pagination from the PageType into  OffsetType
func (p *Pagination) SetLimitOffset() {
	if p.tp == TpPage {
		p.first = p.second
		p.second = p.first * p.second
		p.tp = TpOffset
	}
}

// SetValue sets the pagination value for given parameter
func (p *Pagination) SetValue(val int, parameter Parameter) {
	switch parameter {
	case ParamLimit, ParamNumber:
		p.first = val
	case ParamOffset, ParamSize:
		p.second = val
	}
}

// Type gets the pagination type
func (p *Pagination) Type() Type {
	return p.tp
}

// String implements Stringer interface
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

// GetLimitOffset gets the limit and offset values from the given pagination
func (p *Pagination) GetLimitOffset() (limit, offset int) {
	switch p.tp {
	case TpOffset:
		limit = p.first
		offset = p.second
	case TpPage:
		limit = p.second
		offset = p.second * p.first
	}
	return
}

// GetNumberSize gets the page number and size
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
	if p.tp == TpOffset {
		if p.first < 0 {
			return errors.New("Pagination limit lower than -1")
		}

		if p.second < 0 {
			return errors.New("Pagination offset lower than 0")
		}
		return nil
	}

	if p.first < 0 {
		return errors.New("Pagination page-number lower than 0")
	}

	if p.second < 0 {
		return errors.New("Pagination page-size lower than 0")
	}
	return nil
}
