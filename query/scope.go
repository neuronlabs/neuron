package query

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// Scope is the query's structure that contains information required
// for the processor to operate.
// The scope has its unique 'ID', contains predefined model, operational value, fieldset, filters, sorts and pagination.
// It also contains the mapping of the included scopes.
type Scope struct {
	// id is the unique identification of the scope.
	ID uuid.UUID
	// mStruct is a modelStruct this scope is based on.
	ModelStruct *mapping.ModelStruct
	// Models are the models values used within the context of this query.
	Models []mapping.Model
	// Fieldset represents fieldset defined for the whole scope of this query.
	FieldSet mapping.FieldSet
	// BulkFieldSet are the fieldsets stored for the batch processes. This values are set only when the
	// main fieldset is not defined for the query.
	BulkFieldSets *BulkFieldSet
	// Filters contains all filters for given query.
	Filters Filters
	// SortingOrder are the query sort fields.
	SortingOrder []*SortField
	// IncludedRelations contain fields to include. If the included field is a relationship type, then
	// specific included field contains information about it
	IncludedRelations []*IncludedRelation
	// Pagination is the query pagination.
	Pagination *Pagination
	// Transaction is current scope's transaction.
	Transaction *Transaction

	// store stores the scope's related key values
	store map[interface{}]interface{}
}

// NewScope creates the scope for the provided model with respect to the provided internalController 'c'.
func NewScope(model *mapping.ModelStruct, models ...mapping.Model) *Scope {
	return newQueryScope(model, models...)
}

// Copy creates a copy of the given scope.
func (s *Scope) Copy() *Scope {
	return s.copy()
}

// FormatQuery formats the scope's query into the url.Models.
func (s *Scope) FormatQuery() url.Values {
	return s.formatQuery()
}

// StoreGet gets the value from the scope's Store for given 'key'.
func (s *Scope) StoreGet(key interface{}) (value interface{}, ok bool) {
	value, ok = s.store[key]
	return value, ok
}

// StoreSet sets the 'key' and 'value' in the given scope's store.
func (s *Scope) StoreSet(key, value interface{}) {
	if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s][%s] Store addModel key: '%v', value: '%v'", s.ID, s.ModelStruct.Collection(), key, value)
	}
	s.store[key] = value
}

// String implements fmt.Stringer interface.
func (s *Scope) String() string {
	sb := &strings.Builder{}

	// Scope ID
	sb.WriteString("SCOPE[" + s.ID.String() + "][" + s.ModelStruct.Collection() + "]")

	// Fieldset
	sb.WriteString(" Fieldset")
	if len(s.FieldSet) == len(s.ModelStruct.Fields()) {
		sb.WriteString("(default)")
	}
	sb.WriteString(": [")
	var i int
	for _, field := range s.FieldSet {
		sb.WriteString(field.NeuronName())
		if i != len(s.FieldSet)-1 {
			sb.WriteRune(',')
		}
		i++
	}
	sb.WriteRune(']')

	// Filters
	if len(s.Filters) > 0 {
		sb.WriteString(" Primary Filters: ")
		sb.WriteString(s.Filters.String())
	}

	if s.Pagination != nil {
		sb.WriteString(" Pagination: ")
		sb.WriteString(s.Pagination.String())
	}

	if len(s.SortingOrder) > 0 {
		sb.WriteString(" SortingOrder: ")
		for j, field := range s.SortingOrder {
			sb.WriteString(field.StructField.NeuronName())
			if j != len(s.SortingOrder)-1 {
				sb.WriteRune(',')
			}
		}
	}
	return sb.String()
}

/**

Private scope methods

*/

func (s *Scope) copy() *Scope {
	copiedScope := &Scope{
		ID:          uuid.New(),
		ModelStruct: s.ModelStruct,
		store:       map[interface{}]interface{}{},
	}

	copiedScope.Models = s.Models

	if s.FieldSet != nil {
		copiedScope.FieldSet = make([]*mapping.StructField, len(s.FieldSet))
		for k, v := range s.FieldSet {
			copiedScope.FieldSet[k] = v
		}
	}
	if s.BulkFieldSets != nil {
		copiedScope.BulkFieldSets = &BulkFieldSet{
			FieldSets: make([]mapping.FieldSet, len(s.BulkFieldSets.FieldSets)),
			Indices:   map[string][]int{},
		}
		for i, fieldset := range s.BulkFieldSets.FieldSets {
			copiedScope.BulkFieldSets.FieldSets[i] = fieldset
		}
		for k, v := range s.BulkFieldSets.Indices {
			copiedScope.BulkFieldSets.Indices[k] = v
		}
	}

	if s.Filters != nil {
		copiedScope.Filters = make([]*FilterField, len(s.Filters))
		for i, v := range s.Filters {
			copiedScope.Filters[i] = v.Copy()
		}
	}

	if s.SortingOrder != nil {
		copiedScope.SortingOrder = make([]*SortField, len(s.SortingOrder))
		for i, v := range s.SortingOrder {
			copiedScope.SortingOrder[i] = v.Copy()
		}
	}

	if s.IncludedRelations != nil {
		copiedScope.IncludedRelations = make([]*IncludedRelation, len(s.IncludedRelations))
		for i, v := range s.IncludedRelations {
			copiedScope.IncludedRelations[i] = v.copy()
		}
	}

	return copiedScope
}

func (s *Scope) formatQuery() url.Values {
	q := url.Values{}
	s.formatQueryFilters(q)
	s.formatQuerySorts(q)
	s.formatQueryPagination(q)
	s.formatQueryFieldset(q)
	s.formatQueryIncludes(q)
	return q
}

func (s *Scope) formatQuerySorts(q url.Values) {
	for _, sort := range s.SortingOrder {
		sort.FormatQuery(q)
	}
}

func (s *Scope) formatQueryPagination(q url.Values) {
	if s.Pagination != nil {
		s.Pagination.FormatQuery(q)
	}
}

func (s *Scope) formatQueryFilters(q url.Values) {
	for _, filter := range s.Filters {
		filter.FormatQuery(q)
	}
}

func (s *Scope) formatQueryFieldset(q url.Values) {
	if s.FieldSet != nil {
		fieldsKey := fmt.Sprintf("%s[%s]", ParamFields, s.ModelStruct.Collection())
		var values string
		var i int
		for _, field := range s.FieldSet {
			values += field.NeuronName()
			if i != len(s.FieldSet)-1 {
				values += ","
			}
			i++
		}
		q.Add(fieldsKey, values)
	}
}

func (s *Scope) formatQueryIncludes(q url.Values) {
	var includes []string
	for _, included := range s.IncludedRelations {
		includes = append(includes, included.StructField.NeuronName())
		fieldsKey := fmt.Sprintf("%s[%s]", ParamFields, included.StructField.Relationship().Struct().Collection())
		var values string
		var i int
		for _, field := range included.Fieldset {
			values += field.NeuronName()
			if i != len(s.FieldSet)-1 {
				values += ","
			}
			i++
		}
		q.Add(fieldsKey, values)
	}
	if len(includes) > 0 {
		q.Add("include", strings.Join(includes, ","))
	}
}

func newQueryScope(model *mapping.ModelStruct, models ...mapping.Model) *Scope {
	s := newScope(model)
	s.Models = models
	return s
}

// initialize new scope with added primary field to fieldset
func newScope(modelStruct *mapping.ModelStruct) *Scope {
	s := &Scope{
		ID:          uuid.New(),
		ModelStruct: modelStruct,
		store:       map[interface{}]interface{}{},
	}

	if log.CurrentLevel() <= log.LevelDebug2 {
		log.Debug2f("[SCOPE][%s][%s] query new scope", s.ID.String(), modelStruct.Collection())
	}
	return s
}
