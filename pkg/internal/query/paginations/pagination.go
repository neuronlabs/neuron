package paginations

import (
	"errors"
)

// PaginationType is the enum that describes the type of pagination
type Type int

const (
	TpOffset Type = iota
	TpPage
	TpCursor
)

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
	Type Type
}

// CheckPagination checks if the given pagination is valid
func CheckPagination(p *Pagination) error {
	return p.check()
}

// GetLimitOffset gets the limit and offset values from the given pagination
func (p *Pagination) GetLimitOffset() (limit, offset int) {
	switch p.Type {
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
