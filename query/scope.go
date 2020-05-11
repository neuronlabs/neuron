package query

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

// Scope is the query's structure that contains information required
// for the processor to operate.
// The scope has its unique 'ID', contains predefined model, operational value, fieldset, filters, sorts and pagination.
// It also contains the mapping of the included scopes.
type Scope struct {
	// id is the unique identification of the scope.
	id uuid.UUID
	// DB defines the database interface for given scope.
	db DB
	// mStruct is a modelStruct this scope is based on.
	mStruct *mapping.ModelStruct
	// Models are the models values used within the context of this query.
	Models []mapping.Model
	// Fieldset represents fieldset defined for the whole scope of this query.
	Fieldset mapping.FieldSet
	// ModelsFieldsets are the fieldsets stored for the batch processes. This values are set only when the
	// main fieldset is not defined for the query.
	ModelsFieldsets []mapping.FieldSet
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
	// autoSelectedFields is the flag that defines if the query had automatically selected fieldset.
	autoSelectedFields bool
}

// NewScope creates the scope for the provided model with respect to the provided internalController 'c'.
func NewScope(db DB, model *mapping.ModelStruct) *Scope {
	return newQueryScope(db, model)
}

// Copy creates a copy of the given scope.
func (s *Scope) Copy() *Scope {
	return s.copy()
}

// Count returns the number of all model instances in the repository that matches given query.
func (s *Scope) Count(ctx context.Context) (int64, error) {
	counter, isCounter := s.repository().(Counter)
	if !isCounter {
		return 0, errors.Newf(repository.ClassNotImplements, "repository for model:'%s' doesn't implement Counter interface", s.mStruct)
	}
	s.filterSoftDeleted()
	return counter.Count(ctx, s)
}

// DB gets the scope's DB creator.
func (s *Scope) DB() DB {
	return s.db
}

// Exists returns true or false depending if there are any rows matching the query.
func (s *Scope) Exists(ctx context.Context) (bool, error) {
	exister, isExister := s.repository().(Exister)
	if !isExister {
		return false, errors.Newf(repository.ClassNotImplements, "repository for model: '%s' doesn't implement Exister interface", s.mStruct)
	}
	s.filterSoftDeleted()
	return exister.Exists(ctx, s)
}

// FormatQuery formats the scope's query into the url.Models.
func (s *Scope) FormatQuery() url.Values {
	return s.formatQuery()
}

// Get gets single value from the repository taking into account the scope
// filters and parameters.
func (s *Scope) Get(ctx context.Context) (mapping.Model, error) {
	var err error
	// Check if the pagination is already set.
	if s.Pagination != nil {
		return nil, errors.Newf(ClassInvalidField, "cannot get single model with custom pagination")
	}
	// Assure that the result would be only a single value.
	s.Limit(1)
	results, err := s.Find(ctx)
	if err != nil {
		return nil, err
	}
	// if there is no result return an error of class QueryValueNoResult.
	if len(results) == 0 {
		return nil, errors.New(ClassNoResult, "values not found")
	}
	return results[0], nil
}

// ID returns the scope's identity number stored as the UUID.
func (s *Scope) ID() uuid.UUID {
	return s.id
}

// Struct returns scope's model's structure - *mapping.ModelStruct.
func (s *Scope) Struct() *mapping.ModelStruct {
	return s.mStruct
}

// StoreGet gets the value from the scope's Store for given 'key'.
func (s *Scope) StoreGet(key interface{}) (value interface{}, ok bool) {
	value, ok = s.store[key]
	return value, ok
}

// StoreSet sets the 'key' and 'value' in the given scope's store.
func (s *Scope) StoreSet(key, value interface{}) {
	if log.CurrentLevel().IsAllowed(log.LevelDebug3) {
		log.Debug3f("SCOPE[%s][%s] Store AddModel key: '%v', value: '%v'", s.ID(), s.mStruct.Collection(), key, value)
	}
	s.store[key] = value
}

// String implements fmt.Stringer interface.
func (s *Scope) String() string {
	sb := &strings.Builder{}

	// Scope ID
	sb.WriteString("SCOPE[" + s.ID().String() + "][" + s.Struct().Collection() + "]")

	// Fieldset
	sb.WriteString(" Fieldset")
	if len(s.Fieldset) == len(s.mStruct.Fields()) {
		sb.WriteString("(default)")
	}
	sb.WriteString(": [")
	var i int
	for _, field := range s.Fieldset {
		sb.WriteString(field.NeuronName())
		if i != len(s.Fieldset)-1 {
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
		id:      uuid.New(),
		db:      s.db,
		mStruct: s.mStruct,
		store:   map[interface{}]interface{}{},
	}

	copiedScope.Models = s.Models

	if s.Fieldset != nil {
		copiedScope.Fieldset = make([]*mapping.StructField, len(s.Fieldset))
		for k, v := range s.Fieldset {
			copiedScope.Fieldset[k] = v
		}
	}
	if len(s.ModelsFieldsets) != 0 {
		copiedScope.ModelsFieldsets = make([]mapping.FieldSet, len(s.ModelsFieldsets))
		for i, fieldset := range s.ModelsFieldsets {
			copiedScope.ModelsFieldsets[i] = make(mapping.FieldSet, len(fieldset))
			for j, field := range fieldset {
				copiedScope.ModelsFieldsets[i][j] = field
			}
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
	if s.Fieldset != nil {
		fieldsKey := fmt.Sprintf("%s[%s]", ParamFields, s.Struct().Collection())
		var values string
		var i int
		for _, field := range s.Fieldset {
			values += field.NeuronName()
			if i != len(s.Fieldset)-1 {
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
			if i != len(s.Fieldset)-1 {
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

func (s *Scope) repository() repository.Repository {
	repo, err := s.DB().Controller().GetRepository(s.mStruct)
	if err != nil {
		log.Panicf("Can't find repository for model: %s", s.mStruct.String())
	}
	return repo
}

func newQueryScope(db DB, model *mapping.ModelStruct, models ...mapping.Model) *Scope {
	s := newScope(db, model)
	s.Models = models
	return s
}

// initialize new scope with added primary field to fieldset
func newScope(db DB, modelStruct *mapping.ModelStruct) *Scope {
	s := &Scope{
		id:      uuid.New(),
		db:      db,
		mStruct: modelStruct,
		store:   map[interface{}]interface{}{},
	}

	if log.CurrentLevel() <= log.LevelDebug2 {
		log.Debug2f("[SCOPE][%s][%s] query new scope", s.id.String(), modelStruct.Collection())
	}
	return s
}
