package query

import (
	"github.com/kucjac/jsonapi/pkg/internal/query/filters"
	"github.com/kucjac/jsonapi/pkg/internal/query/scope"
)

// Scope is the Queries heart and soul which keeps all possible information
// Within it's structure
type Scope scope.Scope

// AddFilter adds the given scope's filter field
func (s *Scope) AddFilter(filter *FilterField) error {
	return scope.AddFilterField((*scope.Scope)(s), (*filters.FilterField)(filter))
}

func (s *Scope) AddRawFilter(raw string) error {
	return nil
}

// AttributeFilters returns scope's attribute filters
func (s *Scope) AttributeFilters() []*FilterField {
	var res []*FilterField
	for _, filter := range scope.FiltersAttributes((*scope.Scope)(s)) {

		res = append(res, (*FilterField)(filter))
	}
	return res
}

// ForeignFilters returns scope's foreign key filters
func (s *Scope) ForeignFilters() []*FilterField {
	var res []*FilterField
	for _, filter := range scope.FiltersForeigns((*scope.Scope)(s)) {
		res = append(res, (*FilterField)(filter))
	}
	return res
}

// FilterKeyFilters returns scope's primary filters
func (s *Scope) FilterKeyFilters() []*FilterField {
	var res []*FilterField
	for _, filter := range scope.FiltersKeys((*scope.Scope)(s)) {
		res = append(res, (*FilterField)(filter))
	}
	return res
}

// PrimaryFilters returns scope's primary filters
func (s *Scope) PrimaryFilters() []*FilterField {
	var res []*FilterField
	for _, filter := range scope.FiltersPrimary((*scope.Scope)(s)) {
		res = append(res, (*FilterField)(filter))
	}
	return res
}

// RelationFilters returns scope's relation fields filters
func (s *Scope) RelationFilters() []*FilterField {
	var res []*FilterField
	for _, filter := range scope.FiltersRelationFields((*scope.Scope)(s)) {
		res = append(res, (*FilterField)(filter))
	}
	return res
}

// SetPagination sets the Pagination for the scope.
func (s *Scope) SetPagination(p *Pagination) error {
	return scope.SetPagination((*scope.Scope)(s), p.Pagination)
}
