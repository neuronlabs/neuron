package query

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

// reduceRelationshipFilters reduces all foreign relationship filters into a model's primary field key filters.
func (s *Scope) reduceRelationshipFilters(ctx context.Context) error {
	// TODO: Execute in separate goroutines.
	for _, filter := range s.Filters {
		if !filter.StructField.IsRelationship() {
			continue
		}
		err := s.reduceRelationshipFilter(ctx, filter)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scope) reduceRelationshipFilter(ctx context.Context, filter *FilterField) error {
	switch filter.StructField.Relationship().Kind() {
	case mapping.RelBelongsTo:
		return s.reduceBelongsToRelationshipFilter(ctx, filter)
	case mapping.RelHasMany, mapping.RelHasOne:
		return s.reduceHasManyRelationshipFilter(ctx, filter)
	case mapping.RelMany2Many:
		return s.reduceMany2ManyRelationshipFilter(ctx, filter)
	default:
		return errors.Newf(ClassFilterField, "filter's field: '%s' is not a relationship", filter.StructField)
	}
}

func (s *Scope) reduceBelongsToRelationshipFilter(ctx context.Context, filter *FilterField) error {
	if log.Level().IsAllowed(log.LevelDebug3) {
		log.Debug3f(s.logFormat("reduceBelongsToRelationshipFilter field: '%s'"), filter.StructField)
	}
	// For the belongs to relationships if the filters are only over the relations primary key then we can change
	// the filter from relationship into the foreign key for the scope 's'
	var onlyPrimes = true
	for _, nested := range filter.Nested {
		if nested.StructField.Kind() != mapping.KindPrimary {
			onlyPrimes = false
			break
		}
	}

	if onlyPrimes {
		// Convert the filters into the foreign key filters of the root scope
		foreignKey := filter.StructField.Relationship().ForeignKey()
		filterField := s.getOrCreateFieldFilter(foreignKey)

		for _, nested := range filter.Nested {
			filterField.Values = append(filterField.Values, nested.Values...)
		}
		return nil
	}

	// Otherwise find all primary key values from the relationship's model that matches the filters.
	relatedQuery := s.DB().QueryCtx(ctx, filter.StructField.Relationship().Struct())
	for _, nestedFilter := range filter.Nested {
		relatedQuery.Filter(nestedFilter)
	}

	models, err := relatedQuery.Select(filter.StructField.Relationship().Struct().Primary()).Find()
	if err != nil {
		return err
	}
	if len(models) == 0 {
		log.Debug2f(s.logFormat("no belongs to relationship: '%s' results found"), filter.StructField)
		return errors.Newf(ClassNoResult, "no relationship: '%s' filter results found", filter.StructField)
	}

	// Get primary key values and set as the scope's foreign key field filters.
	primaries := make([]interface{}, len(models))
	for i, model := range models {
		primaries[i] = model.GetPrimaryKeyValue()
	}
	filterField := s.getOrCreateFieldFilter(filter.StructField.Relationship().ForeignKey())
	filterField.Values = append(filterField.Values, OperatorValues{Operator: OpIn, Values: primaries})
	return nil
}

// Has many relationships is a relationship where the foreign model contains the foreign key
// in order to match the given relationship the related scope must be taken with the foreign key in the fieldset
// having the foreign key from the related model, the root scope should have primary field filter key with
// the values of the related model's foreign key results
func (s *Scope) reduceHasManyRelationshipFilter(ctx context.Context, filter *FilterField) error {
	if log.Level() == log.LevelDebug3 {
		log.Debug3f("[SCOPE][%s] reduceHasManyRelationshipFilter field: '%s'", s.ID(), filter.StructField.Name())
	}
	// Check if all the nested filter fields are the foreign key of this relationship. These filters might be
	// primary field filters for the main scope.
	onlyForeign := true
	foreignKey := filter.StructField.Relationship().ForeignKey()
	for _, nested := range filter.Nested {
		if nested.StructField != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// Convert foreign key filter fields into main scope primary key field filters.
		primaryFilter := s.getOrCreateFieldFilter(s.Struct().Primary())
		for _, nested := range filter.Nested {
			primaryFilter.Values = append(primaryFilter.Values, nested.Values...)
		}
		return nil
	}

	relationQuery := s.DB().QueryCtx(ctx, filter.StructField.Relationship().Struct())
	for _, nested := range filter.Nested {
		relationQuery.Filter(nested)
	}
	relationModels, err := relationQuery.Select(foreignKey).Find()
	if err != nil {
		return err
	}
	uniqueForeignKeyValues := make(map[interface{}]struct{})
	for _, relationModel := range relationModels {
		relationFielder, ok := relationModel.(mapping.Fielder)
		if !ok {
			return errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder", filter.StructField.Relationship().Struct())
		}
		// Check if the foreign key field is not zero value.
		isZero, err := relationFielder.IsFieldZero(foreignKey)
		if err != nil {
			return err
		}
		if isZero {
			continue
		}
		foreignKeyValue, err := relationFielder.GetHashableFieldValue(foreignKey)
		if err != nil {
			return err
		}
		uniqueForeignKeyValues[foreignKeyValue] = struct{}{}
	}
	// If there is no foreign key values then no query matches given filter. Return error QueryNoResult.
	if len(uniqueForeignKeyValues) == 0 {
		log.Debug2f(s.logFormat("No results found for the relationship filter for field: %s"), filter.StructField.NeuronName())
		return errors.Newf(ClassNoResult, "no relationship: '%s' filter filterValues found", filter.StructField)
	}
	// Create primary filter that matches all relationship foreign key values.
	primaryFilter := s.getOrCreateFieldFilter(s.mStruct.Primary())
	filterValues := OperatorValues{Operator: OpIn}
	for foreignKeyValue := range uniqueForeignKeyValues {
		filterValues.Values = append(filterValues.Values, foreignKeyValue)
	}
	primaryFilter.Values = append(primaryFilter.Values, filterValues)
	return nil
}

func (s *Scope) reduceMany2ManyRelationshipFilter(ctx context.Context, filter *FilterField) error {
	log.Debug3f(s.logFormat("reduceMany2ManyRelationshipFilter field: '%s'"), filter.StructField)

	// If all the nested filters are the primary key of the related model
	// the only query should be used on the join model's foreign key.
	onlyPrimaries := true
	relatedPrimaryKey := filter.StructField.Relationship().Struct().Primary()
	foreignKey := filter.StructField.Relationship().ForeignKey()

	for _, nested := range filter.Nested {
		if nested.StructField != relatedPrimaryKey && nested.StructField != foreignKey {
			onlyPrimaries = false
			break
		}
	}

	// If only the primaries of the related model were used find only the values from the join table
	joinModel := filter.StructField.Relationship().JoinModel()
	relatedModelForeignKey := filter.StructField.Relationship().ManyToManyForeignKey()
	if onlyPrimaries {
		// Convert the foreign into root scope primary key filter
		joinModelQuery := s.DB().QueryCtx(ctx, joinModel)

		fkFilter := joinModelQuery.Scope().getOrCreateFieldFilter(relatedModelForeignKey)
		for _, nested := range filter.Nested {
			fkFilter.Values = append(fkFilter.Values, nested.Values...)
		}
		// Select only foreign key field - the primary key value for the main scope.
		joinModels, err := joinModelQuery.Select(foreignKey).Find()
		if err != nil {
			return err
		}

		uniqueForeignKeyValues := make(map[interface{}]struct{})
		for _, joinModel := range joinModels {
			relationFielder, ok := joinModel.(mapping.Fielder)
			if !ok {
				return errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder", filter.StructField.Relationship().Struct())
			}
			// Check if the foreign key field is not zero value.
			isZero, err := relationFielder.IsFieldZero(foreignKey)
			if err != nil {
				return err
			}
			if isZero {
				continue
			}
			foreignKeyValue, err := relationFielder.GetHashableFieldValue(foreignKey)
			if err != nil {
				return err
			}
			uniqueForeignKeyValues[foreignKeyValue] = struct{}{}
		}
		// If there is no foreign key values then no query matches given filter. Return error QueryNoResult.
		if len(uniqueForeignKeyValues) == 0 {
			log.Debug2f(s.logFormat("No results found for the relationship filter for field: %s"), filter.StructField.NeuronName())
			return errors.Newf(ClassNoResult, "no relationship: '%s' filter filterValues found", filter.StructField)
		}

		// Create primary filter that matches all relationship foreign key values.
		primaryFilter := s.getOrCreateFieldFilter(s.mStruct.Primary())
		filterValues := OperatorValues{Operator: OpIn}
		for foreignKeyValue := range uniqueForeignKeyValues {
			filterValues.Values = append(filterValues.Values, foreignKeyValue)
		}
		primaryFilter.Values = append(primaryFilter.Values, filterValues)
		return nil
	}

	relationQuery := s.DB().QueryCtx(ctx, filter.StructField.Relationship().Struct())
	for _, nestedFilter := range filter.Nested {
		relationQuery.Filter(nestedFilter)
	}
	relatedModels, err := relationQuery.Select(filter.StructField.Relationship().Struct().Primary()).Find()
	if err != nil {
		return err
	}
	if len(relatedModels) == 0 {
		log.Debug2f(s.logFormat("No relationship: '%s' filter results found"), filter.StructField)
		return errors.Newf(ClassNoResult, "no relationship: '%s' filter results", filter.StructField)
	}
	var primaries []interface{}
	for _, model := range relatedModels {
		primaries = append(primaries, model.GetPrimaryKeyValue())
	}

	joinModels, err := s.DB().QueryCtx(ctx, joinModel).
		Filter(NewFilterField(relatedModelForeignKey, OpIn, primaries...)).
		Select(foreignKey).Find()
	if err != nil {
		return err
	}

	uniqueForeignKeyValues := make(map[interface{}]struct{})
	for _, joinModel := range joinModels {
		relationFielder, ok := joinModel.(mapping.Fielder)
		if !ok {
			return errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder", filter.StructField.Relationship().Struct())
		}
		// Check if the foreign key field is not zero value.
		isZero, err := relationFielder.IsFieldZero(foreignKey)
		if err != nil {
			return err
		}
		if isZero {
			continue
		}
		foreignKeyValue, err := relationFielder.GetHashableFieldValue(foreignKey)
		if err != nil {
			return err
		}
		uniqueForeignKeyValues[foreignKeyValue] = struct{}{}
	}
	// If there is no foreign key values then no query matches given filter. Return error QueryNoResult.
	if len(uniqueForeignKeyValues) == 0 {
		log.Debug2f(s.logFormat("No results found for the relationship filter for field: %s"), filter.StructField.NeuronName())
		return errors.Newf(ClassNoResult, "no relationship: '%s' filter filterValues found", filter.StructField)
	}

	// Create primary filter that matches all relationship foreign key values.
	primaryFilter := s.getOrCreateFieldFilter(s.mStruct.Primary())
	filterValues := OperatorValues{Operator: OpIn}
	for foreignKeyValue := range uniqueForeignKeyValues {
		filterValues.Values = append(filterValues.Values, foreignKeyValue)
	}
	primaryFilter.Values = append(primaryFilter.Values, filterValues)
	return nil
}
