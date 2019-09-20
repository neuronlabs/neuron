package query

import (
	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
)

// Limit sets the maximum number of objects returned by the List process,
// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as omitting the 'Offset' clause.
// Returns error if the scope 's' already has a pagination or the newly created pagination is not valid.
func (s *Scope) Limit(limit, offset int64) error {
	if s.Pagination != nil {
		return errors.NewDet(class.QueryPaginationAlreadySet, "pagination already set")
	}

	s.Pagination = &Pagination{Size: limit, Offset: offset}
	if err := s.Pagination.IsValid(); err != nil {
		return err
	}
	return nil
}

// Page sets the pagination of the type PageNumberSizePagination with the page 'number' and page 'size'.
// Returns error if the pagination is already set for the given scope 's' or the newly created
// pagination is not valid.
func (s *Scope) Page(number, size int64) error {
	if s.Pagination != nil {
		return errors.NewDet(class.QueryPaginationAlreadySet, "pagination already set")
	}

	s.Pagination = &Pagination{Size: size, Offset: number, Type: PageNumberPagination}
	if err := s.Pagination.IsValid(); err != nil {
		return err
	}
	return nil
}
