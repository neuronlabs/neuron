package query

import (
	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
)

// Limit sets the maximum number of objects returned by the List process,
// Offset sets the query result's offset. It says to skip as many object's from the repository
// before beginning to return the result. 'Offset' 0 is the same as omitting the 'Offset' clause.
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

// Page sets the pagination of the type TpPage with the page 'number' and page 'size'.
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
