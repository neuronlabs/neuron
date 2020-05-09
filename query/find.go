package query

import (
	"context"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron/class"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// Find gets the values from the repository taking with respect to the
// query filters, sorts, pagination and included values. Provided context.Context 'ctx'
// would be used while querying the repositories.
func (s *Scope) Find(ctx context.Context) ([]mapping.Model, error) {
	// If no fields were selected - the query searches for all fields.
	if len(s.Fieldset) == 0 {
		s.Fieldset = s.mStruct.Fields()
	}
	// If the query contains any relationship filters, reduce them to this models fields filter.
	if err := s.reduceRelationshipFilters(ctx); err != nil {
		return nil, err
	}
	finder, isFinder := s.repository().(Finder)
	if !isFinder {
		return nil, errors.Newf(class.RepositoryNotImplements, "models: '%s' repository doesn't implement Finder interface", s.mStruct)
	}

	// If the model uses soft delete and the DeletedAt filter is not set yet - filter all models where DeletedAt is null.
	s.filterSoftDeleted()

	// Select foreign key fields for all included belongs to fields.
	// At the end of this function the fieldset wouldn't contain these fields.
	finalFieldset := s.Fieldset
	if len(s.Fieldset) != len(s.mStruct.Fields()) {
		s.selectIncludedBelongsToForeignKeys()
	}

	// Execute find interface.
	if err := finder.Find(ctx, s); err != nil {
		return nil, err
	}
	if len(s.Models) == 0 {
		return s.Models, nil
	}

	// Find all included relationships for given query.
	if err := s.findIncludedRelations(ctx); err != nil {
		return nil, err
	}

	// Execute 'AfterFind' hook if model implements AfterFinder interface.
	for _, model := range s.Models {
		afterFinder, ok := model.(AfterFinder)
		if ok {
			break
		}
		if err := afterFinder.AfterFind(ctx, s.db); err != nil {
			return nil, err
		}
	}
	s.Fieldset = finalFieldset
	return s.Models, nil
}

func (s *Scope) selectIncludedBelongsToForeignKeys() {
	for i := range s.IncludedRelations {
		relationship := s.IncludedRelations[i].StructField.Relationship()
		if relationship.Kind() == mapping.RelBelongsTo {
			if !s.Fieldset.Contains(relationship.ForeignKey()) {
				s.Fieldset = append(s.Fieldset, relationship.ForeignKey())
			}
		}
	}
}

func (s *Scope) filterSoftDeleted() {
	// Check if the model uses soft deletes.
	deletedAt, hasDeletedAt := s.mStruct.DeletedAt()
	if !hasDeletedAt {
		return
	}

	// Don't add soft deletes filters if the scope has any filters for the 'DeletedAt' field.
	for _, filter := range s.Filters {
		if filter.StructField == deletedAt {
			return
		}
	}
	// Where all models where DeletedAt field is null.
	s.Filters = append(s.Filters, NewFilterField(deletedAt, OpIsNull))
}

func (s *Scope) findBelongsToRelation(ctx context.Context, included *IncludedRelation) error {
	// Map foreign key values to related model indexes.
	relationship := included.StructField.Relationship()
	if len(included.Fieldset) == 1 && included.Fieldset.Contains(relationship.Struct().Primary()) {
		return s.findBelongsToRelationShort(included)
	}

	foreignKeyToIndexes := map[interface{}][]int{}
	for i, model := range s.Models {
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return s.errModelNotAFielder()
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
	filterValues := []interface{}{}
	for k := range foreignKeyToIndexes {
		filterValues = append(filterValues, k)
	}
	if len(filterValues) == 0 {
		return nil
	}
	relatedModels, err := s.DB().QueryCtx(ctx, relationship.Struct()).
		Filter(NewFilterField(relationship.Struct().Primary(), OpIn, filterValues...)).
		Select(included.Fieldset...).
		Find()
	if err != nil {
		return err
	}

	for _, relModel := range relatedModels {
		indexes := foreignKeyToIndexes[relModel.GetPrimaryKeyValue()]

		for _, index := range indexes {
			relationer, ok := s.Models[index].(mapping.SingleRelationer)
			if !ok {
				return errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement mapping.SingleRelationer", s.mStruct.String())
			}

			if err = relationer.SetRelationModel(included.StructField, relModel); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Scope) findBelongsToRelationShort(included *IncludedRelation) error {
	// If the only field in the fieldset is relationship primary key - set foreign key value to the relationship primary.
	relationship := included.StructField.Relationship()
	for _, model := range s.Models {
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return errors.Newf(class.ModelNotImplements,
				"model: '%s' does not implement mapping.Fielder interface",
				s.mStruct.String())
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
			return errors.Newf(class.ModelNotImplements,
				"model: '%s' doesn't implement mapping.SingleRelationer",
				s.mStruct.String())
		}
		// Insert new relation model instance, and set it's primary key field value to the 'foreignKeyValue'.
		relationModel := mapping.NewModel(relationship.Struct())
		if err = relationModel.SetPrimaryKeyValue(foreignKeyValue); err != nil {
			return err
		}
		if err = relationer.SetRelationModel(included.StructField, relationModel); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scope) findHasRelation(ctx context.Context, included *IncludedRelation) error {
	relationship := included.StructField.Relationship()
	primaryKeyToIndex := map[interface{}]int{}
	var filterValues []interface{}
	// Iterate over scope models and map their primary key values to their indexes in the scope value slice.
	for i, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			continue
		}

		primaryKeyToIndex[model.GetPrimaryKeyValue()] = i
		filterValues = append(filterValues, model.GetPrimaryKeyValue())
	}

	fieldSet := included.Fieldset
	if !included.Fieldset.Contains(relationship.ForeignKey()) {
		fieldSet = append(fieldSet, relationship.ForeignKey())
	}

	relatedModels, err := s.DB().QueryCtx(ctx, relationship.Struct()).
		Filter(NewFilterField(relationship.ForeignKey(), OpIn, filterValues...)).
		Select(fieldSet...).
		Find()
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
			return s.errModelNotAFielder()
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
				return errors.Newf(class.ModelNotImplements, "model: '%s' does not implement mapping.SingleRelationer interface", s.mStruct.String())
			}
			if err = relationer.SetRelationModel(included.StructField, relatedModel); err != nil {
				return err
			}
		case mapping.RelHasMany:
			relationer, ok := s.Models[index].(mapping.MultiRelationer)
			if !ok {
				return errors.Newf(class.ModelNotImplements, "model: '%s' does not implement mapping.MultiRelationer interface", s.mStruct.String())
			}
			if err = relationer.AddRelationModel(included.StructField, relatedModel); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Scope) findManyToManyRelation(ctx context.Context, included *IncludedRelation) error {
	relationship := included.StructField.Relationship()

	// Map the primary key field value to the slice index. Store the values in the
	filterValues := []interface{}{}
	primariesToIndex := map[interface{}]int{}
	for index, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			continue
		}
		primariesToIndex[model.GetPrimaryKeyValue()] = index
		filterValues = append(filterValues, model.GetPrimaryKeyValue())
	}

	// Get the back reference foreign key - a field that relates to the root models primary key field in the join table.
	backReferenceFK := relationship.ForeignKey()
	// Get the join model foreign key - a field that relates to the relation model primary key in the join table.
	joinModelFK := relationship.ManyToManyForeignKey()

	models, err := s.db.QueryCtx(ctx, relationship.JoinModel()).
		Select(backReferenceFK, joinModelFK).
		Filter(NewFilterField(backReferenceFK, OpIn, filterValues...)).
		Find()
	if err != nil {
		return err
	}
	if len(models) == 0 {
		return nil
	}

	if len(included.Fieldset) == 0 && included.Fieldset[0] == relationship.Struct().Primary() {
		return s.findManyToManyRelationShort(included, models, primariesToIndex)
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
			return errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement Fielder interface", model.NeuronCollectionName())
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
	if !included.Fieldset.Contains(relationship.Struct().Primary()) {
		fieldSet = append(fieldSet, relationship.Struct().Primary())
	}

	relationModels, err := s.db.QueryCtx(ctx, relationship.Struct()).
		Filter(NewFilterField(relationship.Struct().Primary(), OpIn, relationPrimaries...)).
		Select(fieldSet...).
		Find()
	if err != nil {
		return err
	}

	for _, model := range relationModels {
		if model.IsPrimaryKeyZero() {
			continue
		}
		primaryKeyValue := model.GetPrimaryKeyValue()
		indexes := relationPrimariesToIndexes[primaryKeyValue]
		for _, index := range indexes {
			relationer, ok := s.Models[index].(mapping.MultiRelationer)
			if !ok {
				return errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement interface MultiRelationer", s.mStruct)
			}
			if err = relationer.AddRelationModel(included.StructField, model); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Scope) findManyToManyRelationShort(included *IncludedRelation, models []mapping.Model, primariesToIndex map[interface{}]int) error {
	relationship := included.StructField.Relationship()
	for _, model := range models {
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement Fielder interface", model.NeuronCollectionName())
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
		relationModel := mapping.NewModel(relationship.Struct())
		if err = relationModel.SetPrimaryKeyValue(relationPrimaryKey); err != nil {
			return err
		}

		rootRelationer, ok := s.Models[index].(mapping.MultiRelationer)
		if !ok {
			return errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement MultiRelationer", s.mStruct.String())
		}
		if err = rootRelationer.AddRelationModel(included.StructField, relationModel); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scope) errModelNotAFielder() errors.ClassError {
	return errors.Newf(class.ModelNotImplements, "model: '%s' doesn't implement Fielder interface", s.mStruct)
}
