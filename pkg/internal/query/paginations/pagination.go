package paginations

import (
	"errors"
	"fmt"
	"github.com/kucjac/jsonapi/pkg/config"
)

// PaginationType is the enum that describes the type of pagination
type Type int

const (
	TpOffset Type = iota
	TpPage
	TpCursor
)

func (t Type) String() string {
	switch t {
	case TpOffset:
		return "Offset"
	case TpPage:
		return "Page"
	case TpCursor:
		return "Cursor"
	default:
		return "Unknown"
	}
}

// Parameter Defines the given paginate paramater type
type Parameter int

const (
	ParamLimit Parameter = iota
	ParamOffset
	ParamNumber
	ParamSize
)

// Pagination is a struct that keeps variables used to paginate the result
type Pagination struct {
	Limit      int
	Offset     int
	PageNumber int
	PageSize   int

	UseTotal bool
	Total    int

	// Describes which pagination type to use.
	tp Type
}

func NewFromConfig(p *config.Pagination) *Pagination {
	pg := &Pagination{
		Limit:      p.Limit,
		Offset:     p.Offset,
		PageNumber: p.PageNumber,
		PageSize:   p.PageSize,
	}

	return pg
}

// CheckPagination checks if the given pagination is valid
func CheckPagination(p *Pagination) error {
	return p.check()
}

// Check checks if the given pagination is valid
func (p *Pagination) Check() error {
	return p.check()
}

// SetType sets the pagination type
func (p *Pagination) SetType(tp Type) {
	p.tp = tp
}

// Type gets the pagination type
func (p *Pagination) Type() Type {
	return p.tp
}

// String implements Stringer interface
func (p *Pagination) String() string {
	switch p.tp {
	case TpOffset:
		return fmt.Sprintf("Offset Pagination. Limit: %d, Offset: %d", p.Limit, p.Offset)
	case TpPage:
		return fmt.Sprintf("PageType Pagination. PageSize: %d, PageNumber: %d", p.PageSize, p.PageNumber)
	case TpCursor:
		return fmt.Sprintf("Cursor Pagination. Limit: %d, Offset: %d", p.Limit, p.Offset)
	default:
		return "Unknown Pagination"
	}
}

// GetLimitOffset gets the limit and offset values from the given pagination
func (p *Pagination) GetLimitOffset() (limit, offset int) {
	switch p.tp {
	case TpOffset:
		limit = p.Limit
		offset = p.Offset
	case TpPage:
		limit = p.PageSize
		offset = p.PageNumber * p.PageSize
	case TpCursor:
		// not implemented yet
	}
	return
}

func (p *Pagination) check() error {
	var offsetBased, pageBased bool
	if p.Limit != 0 || p.Offset != 0 {
		offsetBased = true
	}
	if p.PageNumber != 0 || p.PageSize != 0 {
		pageBased = true
	}

	if offsetBased && pageBased {
		err := errors.New("Both offset-based and page-based pagination are set.")
		return err
	}
	return nil
}
