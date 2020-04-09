package query

import (
	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
)

// Limit sets the maximum number of objects returned by the List process,
// Returns error if the given scope has already different type of pagination.
func (s *Scope) Limit(limit int64) error {
	if s.Pagination == nil {
		s.Pagination = &Pagination{}
	}
	if s.Pagination.Type != LimitOffsetPagination {
		return errors.NewDet(class.QueryPaginationType, "setting offset on page number based pagination")
	}
	s.Pagination.Size = limit
	return nil
}

// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as omitting the 'Offset' clause.
// Returns error if the given scope has already different type of pagination.
func (s *Scope) Offset(offset int64) error {
	if s.Pagination == nil {
		s.Pagination = &Pagination{}
	}
	if s.Pagination.Type != LimitOffsetPagination {
		return errors.NewDet(class.QueryPaginationType, "setting offset on page number based pagination")
	}
	s.Pagination.Offset = offset
	return nil
}

// PageSize defines pagination page size - maximum amount of returned objects.
// Returns error if given scope already has a pagination of LimitOffsetPagination type.
func (s *Scope) PageSize(size int64) error {
	if s.Pagination == nil {
		s.Pagination = &Pagination{Type: PageNumberPagination}
	}
	if s.Pagination.Type != PageNumberPagination {
		return errors.NewDet(class.QueryPaginationType, "setting offset on page number based pagination")
	}
	s.Pagination.Size = size
	return nil
}

// PageNumber defines the pagination page number.
// Returns error if given scope already has a pagination of LimitOffsetPagination type.
func (s *Scope) PageNumber(number int64) error {
	if s.Pagination == nil {
		s.Pagination = &Pagination{Type: PageNumberPagination}
	}
	if s.Pagination.Type != PageNumberPagination {
		return errors.NewDet(class.QueryPaginationType, "setting offset on page number based pagination")
	}
	s.Pagination.Offset = number
	return nil
}
