package database

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// queryFind gets the values from the repository with respect to the query filters, sorts, pagination and included values.
// Provided 'ctx' context.Context would be used while querying the repositories.
func queryFind(ctx context.Context, db DB, s *query.Scope) ([]mapping.Model, error) {
	if len(s.Models) > 0 {
		if err := refreshQuery(ctx, db, s); err != nil {
			return nil, err
		}
		return s.Models, nil
	}
	// If no fields were selected - the query searches for all fields.
	switch len(s.FieldSets) {
	case 0:
		s.FieldSets = append(s.FieldSets, s.ModelStruct.Fields())
	case 1:
		if len(s.FieldSets[0]) == 0 {
			s.FieldSets[0] = s.ModelStruct.Fields()
		}
	default:
		return nil, errors.WrapDetf(query.ErrInvalidFieldSet, "provided too many fieldsets for the find query")
	}
	// If the query contains any relationship filters, reduce them to this models fields filter.
	if err := reduceRelationshipFilters(ctx, db, s); err != nil {
		return nil, err
	}
	// If the model uses soft delete and the DeletedAt filter is not set yet - filter all models where DeletedAt is null.
	filterSoftDeleted(s)

	// Select foreign key fields for all included belongs to fields.
	// At the end of this function the fieldset wouldn't contain these fields.
	if len(s.FieldSets[0]) != len(s.ModelStruct.Fields()) {
		selectIncludedBelongsToForeignKeys(s)
	}

	// Execute find interface.
	if err := getRepository(db.Controller(), s).Find(ctx, s); err != nil {
		return nil, err
	}
	if len(s.Models) == 0 {
		return s.Models, nil
	}

	// Find all included relationships for given query.
	if err := findIncludedRelations(ctx, db, s); err != nil {
		return nil, err
	}

	// Execute 'AfterFind' hook if model implements AfterFinder interface.
	for _, model := range s.Models {
		afterFinder, ok := model.(AfterFinder)
		if !ok {
			break
		}
		if err := afterFinder.AfterFind(ctx, db); err != nil {
			return nil, err
		}
	}
	return s.Models, nil
}

// queryGet gets single value from the repository taking into account the scope filters and parameters.
func queryGet(ctx context.Context, db DB, s *query.Scope) (mapping.Model, error) {
	var err error
	// Check if the pagination is already set.
	if s.Pagination != nil && (s.Pagination.Limit != 1 || s.Pagination.Offset != 0) {
		return nil, errors.Wrapf(query.ErrInvalidField, "cannot get single model with custom pagination")
	}
	// Assure that the result would be only a single value.
	s.Limit(1)
	results, err := queryFind(ctx, db, s)
	if err != nil {
		return nil, err
	}
	// if there is no result return an error of class QueryValueNoResult.
	if len(results) == 0 {
		return nil, errors.Wrap(query.ErrQueryNoResult, "model not found")
	}
	return results[0], nil
}

func selectIncludedBelongsToForeignKeys(s *query.Scope) {
	for i := range s.IncludedRelations {
		relationship := s.IncludedRelations[i].StructField.Relationship()
		if relationship.Kind() == mapping.RelBelongsTo {
			if !s.FieldSets[0].Contains(relationship.ForeignKey()) {
				s.FieldSets[0] = append(s.FieldSets[0], relationship.ForeignKey())
			}
		}
	}
}

func filterSoftDeleted(s *query.Scope) {
	// Check if the model uses soft deletes.
	deletedAt, hasDeletedAt := s.ModelStruct.DeletedAt()
	if !hasDeletedAt {
		return
	}

	// Don't add soft deletes filters if the scope has any filters for the 'DeletedAt' field.
	for _, f := range s.Filters {
		if isDeletedAtField(deletedAt, f) {
			return
		}
	}
	// Where all models where DeletedAt field is null.
	s.Filters = append(s.Filters, filter.New(deletedAt, filter.OpIsNull))
}

func isDeletedAtField(sField *mapping.StructField, f filter.Filter) bool {
	switch ft := f.(type) {
	case filter.Simple:
		if ft.StructField == sField {
			return true
		}
	case filter.OrGroup:
		for _, sub := range ft {
			if isDeletedAtField(sField, sub) {
				return true
			}
		}
	}
	return false
}

func findBelongsToRelation(ctx context.Context, db DB, s *query.Scope, included *query.IncludedRelation) error {
	// Map foreign key values to related model indexes.
	relationship := included.StructField.Relationship()
	if len(included.Fieldset) == 1 && included.Fieldset.Contains(relationship.RelatedModelStruct().Primary()) {
		return findBelongsToRelationShort(s, included)
	}

	foreignKeyToIndexes := map[interface{}][]int{}
	for i, model := range s.Models {
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return errModelNotAFielder(s)
		}
		// Check if the foreign key value is not 'zero'.
		isZero, err := fielder.IsFieldZero(relationship.ForeignKey())
		if err != nil {
			return err
		}
		if isZero {
			continue
		}

		foreignKeyValue, err := fielder.GetHashableFieldValue(relationship.ForeignKey())
		if err != nil {
			return err
		}
		foreignKeyToIndexes[foreignKeyValue] = append(foreignKeyToIndexes[foreignKeyValue], i)
	}

	// Get unique foreign key filter values.
	var filterValues []interface{}
	for k := range foreignKeyToIndexes {
		filterValues = append(filterValues, k)
	}
	if len(filterValues) == 0 {
		return nil
	}
	q := db.QueryCtx(ctx, relationship.RelatedModelStruct()).
		Filter(filter.New(relationship.RelatedModelStruct().Primary(), filter.OpIn, filterValues...)).
		Select(included.Fieldset...)
	if len(included.IncludedRelations) != 0 {
		q.Scope().IncludedRelations = included.IncludedRelations
	}
	relatedModels, err := q.Find()
	if err != nil {
		return err
	}

	for _, relModel := range relatedModels {
		indexes := foreignKeyToIndexes[relModel.GetPrimaryKeyHashableValue()]

		for _, index := range indexes {
			relationer, ok := s.Models[index].(mapping.SingleRelationer)
			if !ok {
				return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement mapping.SingleRelationer", s.ModelStruct.String())
			}

			if err = relationer.SetRelationModel(included.StructField, relModel); err != nil {
				return err
			}
		}
	}
	return nil
}

func findBelongsToRelationShort(s *query.Scope, included *query.IncludedRelation) error {
	// If the only field in the fieldset is relationship primary key - set foreign key value to the relationship primary.
	relationship := included.StructField.Relationship()
	for _, model := range s.Models {
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return errors.Wrapf(mapping.ErrModelNotImplements,
				"model: '%s' does not implement mapping.Fielder interface",
				s.ModelStruct.String())
		}
		// Check if the foreign key value is not 'zero'.
		isZero, err := fielder.IsFieldZero(relationship.ForeignKey())
		if err != nil {
			return err
		}
		if isZero {
			continue
		}

		foreignKeyValue, err := fielder.GetHashableFieldValue(relationship.ForeignKey())
		if err != nil {
			return err
		}

		relationer, ok := model.(mapping.SingleRelationer)
		if !ok {
			return errors.Wrapf(mapping.ErrModelNotImplements,
				"model: '%s' doesn't implement mapping.SingleRelationer",
				s.ModelStruct.String())
		}
		// Insert new relation model instance, and set it's primary key field value to the 'foreignKeyValue'.
		relationModel := mapping.NewModel(relationship.RelatedModelStruct())
		if err = relationModel.SetPrimaryKeyValue(foreignKeyValue); err != nil {
			return err
		}
		if err = relationer.SetRelationModel(included.StructField, relationModel); err != nil {
			return err
		}

	}
	return nil
}

func findHasRelation(ctx context.Context, db DB, s *query.Scope, included *query.IncludedRelation) error {
	relationship := included.StructField.Relationship()
	primaryKeyToIndex := map[interface{}]int{}
	var filterValues []interface{}
	// Iterate over scope models and map their primary key values to their indexes in the scope value slice.
	for i, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			continue
		}

		primaryKeyToIndex[model.GetPrimaryKeyHashableValue()] = i
		filterValues = append(filterValues, model.GetPrimaryKeyValue())
	}

	fieldSet := included.Fieldset
	if !included.Fieldset.Contains(relationship.ForeignKey()) {
		fieldSet = append(fieldSet, relationship.ForeignKey())
	}

	q := db.QueryCtx(ctx, relationship.RelatedModelStruct()).
		Filter(filter.New(relationship.ForeignKey(), filter.OpIn, filterValues...)).
		Select(fieldSet...)
	if len(included.IncludedRelations) != 0 {
		s := q.Scope()
		s.IncludedRelations = included.IncludedRelations
	}
	relatedModels, err := q.Find()
	if err != nil {
		return err
	}
	if len(relatedModels) == 0 {
		return nil
	}

	// For each related model found set it's value to the main scope models by mapping their
	// foreign key values with main models primaries.
	for _, relatedModel := range relatedModels {
		fielder, ok := relatedModel.(mapping.Fielder)
		if !ok {
			return errModelNotAFielder(s)
		}
		foreignKeyValue, err := fielder.GetHashableFieldValue(relationship.ForeignKey())
		if err != nil {
			return err
		}
		// Get the index of the main model based on the foreign key field value.
		index, ok := primaryKeyToIndex[foreignKeyValue]
		if !ok {
			log.Debugf("Relationship HasMany Foreign Key not found in the primaries map.")
			continue
		}
		switch relationship.Kind() {
		case mapping.RelHasOne:
			relationer, ok := s.Models[index].(mapping.SingleRelationer)
			if !ok {
				return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' does not implement mapping.SingleRelationer interface", s.ModelStruct.String())
			}
			if err = relationer.SetRelationModel(included.StructField, relatedModel); err != nil {
				return err
			}
		case mapping.RelHasMany:
			relationer, ok := s.Models[index].(mapping.MultiRelationer)
			if !ok {
				return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' does not implement mapping.MultiRelationer interface", s.ModelStruct.String())
			}
			if err = relationer.AddRelationModel(included.StructField, relatedModel); err != nil {
				return err
			}
		}
	}
	return nil
}

func findManyToManyRelation(ctx context.Context, db DB, s *query.Scope, included *query.IncludedRelation) error {
	relationship := included.StructField.Relationship()

	// Map the primary key field value to the slice index. Store the values in the
	var filterValues []interface{}
	primariesToIndex := map[interface{}]int{}
	for index, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			continue
		}
		primariesToIndex[model.GetPrimaryKeyHashableValue()] = index
		filterValues = append(filterValues, model.GetPrimaryKeyValue())
	}

	// Get the back reference foreign key - a field that relates to the root models primary key field in the join table.
	backReferenceFK := relationship.ForeignKey()
	// Get the join model foreign key - a field that relates to the relation model primary key in the join table.
	joinModelFK := relationship.ManyToManyForeignKey()

	models, err := db.QueryCtx(ctx, relationship.JoinModel()).
		Select(backReferenceFK, joinModelFK).
		Filter(filter.New(backReferenceFK, filter.OpIn, filterValues...)).
		Find()
	if err != nil {
		return err
	}
	if len(models) == 0 {
		return nil
	}

	if len(included.Fieldset) == 1 && included.Fieldset[0] == relationship.RelatedModelStruct().Primary() && len(included.IncludedRelations) == 0 {
		return findManyToManyRelationShort(s, included, models, primariesToIndex)
	}

	relationPrimariesToIndexes := map[interface{}][]int{}
	// If the field was included the relations needs to find with new scope.
	// Iterate over relatedScope Models
	for index, model := range models {
		if model == nil {
			continue
		}
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder interface", model.NeuronCollectionName())
		}
		isZero, err := fielder.IsFieldZero(backReferenceFK)
		if err != nil {
			return err
		}
		if isZero {
			continue
		}
		foreignKeyValue, err := fielder.GetHashableFieldValue(backReferenceFK)
		if err != nil {
			return err
		}
		relationPrimariesToIndexes[foreignKeyValue] = append(relationPrimariesToIndexes[foreignKeyValue], index)
	}

	// create a scope for the relation
	var relationPrimaries []interface{}
	for k := range relationPrimariesToIndexes {
		relationPrimaries = append(relationPrimaries, k)
	}

	fieldSet := included.Fieldset
	if !included.Fieldset.Contains(relationship.RelatedModelStruct().Primary()) {
		fieldSet = append(fieldSet, relationship.RelatedModelStruct().Primary())
	}
	q := db.QueryCtx(ctx, relationship.RelatedModelStruct()).
		Filter(filter.New(relationship.RelatedModelStruct().Primary(), filter.OpIn, relationPrimaries...)).
		Select(fieldSet...)
	if len(included.IncludedRelations) != 0 {
		q.Scope().IncludedRelations = included.IncludedRelations
	}
	relationModels, err := q.Find()
	if err != nil {
		return err
	}

	for _, model := range relationModels {
		if model.IsPrimaryKeyZero() {
			continue
		}
		primaryKeyValue := model.GetPrimaryKeyHashableValue()
		indexes := relationPrimariesToIndexes[primaryKeyValue]
		for _, index := range indexes {
			relationer, ok := s.Models[index].(mapping.MultiRelationer)
			if !ok {
				return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement interface MultiRelationer", s.ModelStruct)
			}
			if err = relationer.AddRelationModel(included.StructField, model); err != nil {
				return err
			}
		}
	}
	return nil
}

func findManyToManyRelationShort(s *query.Scope, included *query.IncludedRelation, models []mapping.Model, primariesToIndex map[interface{}]int) error {
	relationship := included.StructField.Relationship()
	for _, model := range models {
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder interface", model.NeuronCollectionName())
		}
		// Find related model value from the root scope by matching it's primary key with the back reference foreign key.
		foreignKeyValue, err := fielder.GetHashableFieldValue(relationship.ForeignKey())
		if err != nil {
			return err
		}
		index, ok := primariesToIndex[foreignKeyValue]
		if !ok {
			log.Errorf("Get Many2Many relationship. Can't find primary field within the mapping, for ID: %v", foreignKeyValue)
			continue
		}
		// Get primary key value for the new relation model - from join scope mode
		relationPrimaryKey, err := fielder.GetHashableFieldValue(relationship.ManyToManyForeignKey())
		if err != nil {
			return err
		}
		// Insert a new model instance and add set it's primary key value.
		relationModel := mapping.NewModel(relationship.RelatedModelStruct())
		if err = relationModel.SetPrimaryKeyValue(relationPrimaryKey); err != nil {
			return err
		}

		rootRelationer, ok := s.Models[index].(mapping.MultiRelationer)
		if !ok {
			return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement MultiRelationer", s.ModelStruct.String())
		}
		if err = rootRelationer.AddRelationModel(included.StructField, relationModel); err != nil {
			return err
		}
	}
	return nil
}

func errModelNotAFielder(s *query.Scope) error {
	return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.ModelStruct)
}
