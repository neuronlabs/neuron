package jsonapi

// PaginationType is the enum that describes the type of pagination
type PaginationType int

const (
	OffsetPaginate PaginationType = iota
	PagePaginate
	CursorPaginate
)

type Pagination struct {
	Limit      int
	Offset     int
	PageNumber int
	PageSize   int

	UseTotal bool
	Total    int

	// Describes which pagination type to use.
	Type PaginationType
}

func (p *Pagination) GetLimitOffset() (limit, offset int) {
	switch p.Type {
	case OffsetPaginate:
		limit = p.Limit
		offset = p.Offset
	case PagePaginate:
		limit = p.PageSize
		offset = p.PageNumber * p.PageSize
	case CursorPaginate:
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
