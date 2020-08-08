package database

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// AddRelations appends relationship 'relField' 'relationModels' to the scope values.
func queryAddRelations(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField, relationModels ...mapping.Model) error {
	if s.ModelStruct != relField.Struct() {
		return errors.Wrapf(query.ErrInvalidField, "provided relation field: '%s' doesn't belong to the scope's relationModel: '%s'", relField.String(), s.ModelStruct.String())
	}
	switch relField.Relationship().Kind() {
	case mapping.RelHasMany:
		return addRelationHasMany(ctx, db, s, relField, relationModels)
	case mapping.RelMany2Many:
		return addRelationMany2Many(ctx, db, s, relField, relationModels)
	case mapping.RelHasOne:
		return addRelationHasOne(ctx, db, s, relField, relationModels)
	case mapping.RelBelongsTo:
		return setBelongsToRelation(ctx, db, s, relField, relationModels)
	default:
		return errors.Wrapf(query.ErrInvalidField, "provided relation field: '%s' with invalid relationship kind: '%s'", relField.String(), relField.Relationship().Kind().String())
	}
}

func addRelationMany2Many(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField, relationModels []mapping.Model) error {
	var (
		joinModels []mapping.Model
		err        error
	)
	relationship := relField.Relationship()

	// Check and set manually timestamps if the join model contains such values.
	// This function insert new models with selected fields, thus automatic timestamps settings doesn't work.
	createdAt, hasCreatedAt := relationship.JoinModel().CreatedAt()
	updatedAt, hasUpdatedAt := relationship.JoinModel().UpdatedAt()
	nowTS := db.Controller().Now()

	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return errors.Wrap(query.ErrInvalidModels, "one of provided model's primary key value has zero value")
		}

		// For each related model insert join model instance with the foreign key value from 'model' primary key
		// and many2many foreign key value from relationModel primary key.
		for _, relationModel := range relationModels {
			if relationModel.IsPrimaryKeyZero() {
				return errors.Wrap(query.ErrInvalidModels, "one of provided relation model's primary key value has zero value")
			}

			joinModel := mapping.NewModel(relationship.JoinModel())
			joinModelFielder, ok := joinModel.(mapping.Fielder)
			if !ok {
				return errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder interface", joinModel.NeuronCollectionName())
			}
			err = joinModelFielder.SetFieldValue(relationship.ForeignKey(), model.GetPrimaryKeyHashableValue())
			if err != nil {
				return err
			}
			err = joinModelFielder.SetFieldValue(relationship.ManyToManyForeignKey(), relationModel.GetPrimaryKeyHashableValue())
			if err != nil {
				return err
			}

			if hasCreatedAt {
				if err = joinModelFielder.SetFieldValue(createdAt, nowTS); err != nil {
					return err
				}
			}
			if hasUpdatedAt {
				if err = joinModelFielder.SetFieldValue(updatedAt, nowTS); err != nil {
					return err
				}
			}
			joinModels = append(joinModels, joinModel)
		}
	}

	fieldSet := mapping.FieldSet{relationship.JoinModel().Primary(), relationship.ForeignKey(), relationship.ManyToManyForeignKey()}
	if hasCreatedAt {
		fieldSet = append(fieldSet, createdAt)
	}
	if hasUpdatedAt {
		fieldSet = append(fieldSet, updatedAt)
	}
	return db.QueryCtx(ctx, relationship.JoinModel(), joinModels...).
		Select(fieldSet...).
		Insert()
}

func addRelationHasMany(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField, relationModels []mapping.Model) error {
	// The nature of the HasMany relationship doesn't allow to set multiple the same
	// relation models to multiple model values. In the belongs to relation field there must be only a single foreign key.
	if len(s.Models) > 1 {
		return errors.Wrapf(query.ErrInvalidModels, "relation field: '%s' of HasMany kind cannot be added to multiple model values", relField.String())
	}
	// Iterate over relationship models and set relationship foreign key value to the
	// scope's primary key value.
	model := s.Models[0]
	if model == nil {
		return errors.Wrapf(query.ErrInvalidModels, "provided nil model value")
	}
	// Check if the model's primary key is non zero.
	if model.IsPrimaryKeyZero() {
		return errors.Wrap(query.ErrInvalidInput, "provided model's primary key value has zero value")
	}
	relationship := relField.Relationship()
	updatedAt, hasUpdatedAt := relationship.RelatedModelStruct().UpdatedAt()
	updatedAtTS := db.Controller().Now()
	// Iterate over all relation models and set relation's foreign key to the primary key of the 'model'.
	for _, relationModel := range relationModels {
		if relationModel.IsPrimaryKeyZero() {
			return errors.Wrapf(query.ErrInvalidInput, "relation model value has zero primary key value")
		}
		// addModel the foreign key field to the model's primary key value.
		fielder, ok := relationModel.(mapping.Fielder)
		if !ok {
			return errModelNotImplements(relationship.RelatedModelStruct(), "Fielder")
		}
		err := fielder.SetFieldValue(relationship.ForeignKey(), model.GetPrimaryKeyHashableValue())
		if err != nil {
			return err
		}

		// If the model has UpdatedAt timestamp we need to set it manually as we would query update with selected fields.
		if hasUpdatedAt {
			if err = fielder.SetFieldValue(updatedAt, updatedAtTS); err != nil {
				return err
			}
		}
	}
	// update all relation models using their primary key as well as this relationship foreign key.
	_, err := db.
		QueryCtx(ctx, relationship.RelatedModelStruct(), relationModels...).
		Select(fieldSetWithUpdatedAt(relationship.RelatedModelStruct(), relationship.RelatedModelStruct().Primary(), relationship.ForeignKey())...).
		Update()
	return err
}

func addRelationHasOne(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) error {
	if len(relationModels) > 1 {
		return errors.Wrap(query.ErrInvalidInput, "cannot set multiple relation models to the models with 'has one' relationship")
	}
	relationModel := relationModels[0]
	if relationModel.IsPrimaryKeyZero() {
		return errors.Wrap(query.ErrInvalidInput, "provided relation has zero value primary key")
	}
	if len(s.Models) > 1 {
		return errors.Wrap(query.ErrInvalidInput, "cannot set multiple models with 'has one' relation ship to the single relation field")
	}

	model := s.Models[0]
	if model.IsPrimaryKeyZero() {
		return errors.Wrap(query.ErrInvalidInput, "provide model has zero value primary key")
	}

	relationship := relationField.Relationship()

	fielder, ok := relationModel.(mapping.Fielder)
	if !ok {
		return errModelNotImplements(relationship.RelatedModelStruct(), "Fielder")
	}
	if err := fielder.SetFieldValue(relationship.ForeignKey(), model.GetPrimaryKeyHashableValue()); err != nil {
		return err
	}

	fieldSet := mapping.FieldSet{relationship.RelatedModelStruct().Primary(), relationship.ForeignKey()}

	updatedAt, hasUpdatedAt := relationship.RelatedModelStruct().UpdatedAt()
	if hasUpdatedAt {
		if err := fielder.SetFieldValue(updatedAt, db.Controller().Now()); err != nil {
			return err
		}
		fieldSet = append(fieldSet, updatedAt)
	}

	_, err := db.QueryCtx(ctx, relationship.RelatedModelStruct(), relationModel).
		Select(fieldSet...).
		Update()
	return err
}

//
// Clear relations.
//

// queryClearRelations clears the 'relation' for provided query scope models.
func queryClearRelations(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField) (int64, error) {
	if s.ModelStruct != relField.Struct() {
		return 0, errors.Wrapf(query.ErrInvalidField, "provided relation field: '%s' doesn't belong to the scope's relationModel: '%s'", relField.String(), s.ModelStruct.String())
	}
	if !relField.IsRelationship() {
		return 0, errors.Wrapf(query.ErrInvalidField, "provided field: '%s' is not a relationship", relField)
	}
	if err := requireNoFilters(s); err != nil {
		return 0, err
	}
	if len(s.Models) == 0 {
		return 0, errors.Wrapf(query.ErrNoModels, "provided no models to remove the relations")
	}
	switch relField.Relationship().Kind() {
	case mapping.RelMany2Many:
		return clearMany2ManyRelations(ctx, db, s, relField)
	case mapping.RelHasOne:
		return clearHasOneRelation(ctx, db, s, relField)
	case mapping.RelHasMany:
		return clearHasManyRelations(ctx, db, s, relField)
	case mapping.RelBelongsTo:
		return clearBelongsToRelation(ctx, db, s, relField)
	default:
		return 0, errors.Wrapf(query.ErrInternal, "invalid relationship kind: '%s'", relField.Relationship().Kind())
	}
}

func clearBelongsToRelation(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField) (int64, error) {
	// Removing relations from the belongs to relationship involves setting foreign key field's value to nullable.
	if !relField.Relationship().ForeignKey().IsPtr() {
		return 0, errors.Wrapf(mapping.ErrFieldNotNullable, "provided relation field: '%s' doesn't allow setting null values", relField)
	}
	if err := requireNoFilters(s); err != nil {
		return 0, err
	}
	// Iterate over scope models and set relationship foreign key value to zero - nullable.
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.Wrapf(query.ErrInvalidModels, "one of the model values has zero primary key value")
		}
		fielder := model.(mapping.Fielder)
		if err := fielder.SetFieldZeroValue(relField.Relationship().ForeignKey()); err != nil {
			return 0, err
		}
	}
	// make a copy of given query.
	s = s.Copy()
	// addModel only two fields in the fieldset.
	s.FieldSets = []mapping.FieldSet{fieldSetWithUpdatedAt(s.ModelStruct, s.ModelStruct.Primary(), relField.Relationship().ForeignKey())}
	return queryUpdate(ctx, db, s)
}

func clearHasOneRelation(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField) (int64, error) {
	// Removing relations from the HasOne relationships involves setting related foreign key value to zero.
	if !relField.Relationship().ForeignKey().IsPtr() {
		return 0, errors.Wrapf(mapping.ErrFieldNotNullable, "provided relation field: '%s' doesn't allow setting null values", relField)
	}
	// If the relationship structure implements update hooks we need to firstly get he
	// related model. The user might set something in the hooks.
	// Otherwise we can change it by using query update.
	relationer := s.Models[0].(mapping.SingleRelationer)
	relationModel, err := relationer.GetRelationModel(relField)
	if err != nil {
		return 0, err
	}
	// Check if the related model implements update hooks.
	_, gettingRelationRequired := relationModel.(BeforeUpdater)
	if !gettingRelationRequired {
		_, gettingRelationRequired = relationModel.(AfterUpdater)
	}

	// Get all relationship models where foreign key value is equal to 'models' primary key.
	var primaryKeyValues []interface{}
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.Wrapf(query.ErrInvalidModels, "one of the model values has zero primary key value")
		}
		primaryKeyValues = append(primaryKeyValues, model.GetPrimaryKeyHashableValue())
	}
	relationModelStruct := relField.Relationship().RelatedModelStruct()
	if gettingRelationRequired {
		relationModels, err := db.QueryCtx(ctx, relationModelStruct).
			Select(relationModelStruct.Primary()).
			Filter(filter.New(relField.Relationship().ForeignKey(), filter.OpIn, primaryKeyValues...)).
			Find()
		if err != nil {
			return 0, err
		}

		// update relationship models using their primary and foreign key fields.
		// The foreign key has it's own zero value - previous query pulled only
		return db.QueryCtx(ctx, relationModelStruct, relationModels...).
			Select(fieldSetWithUpdatedAt(relationModelStruct, relationModelStruct.Primary(), relField.Relationship().ForeignKey())...).
			Update()
	}
	// update using queryUpdate - set foreign key values to zero.
	return db.QueryCtx(ctx, relationModelStruct, mapping.NewModel(relationModelStruct)).
		// Select a foreign key - it must be in a zero like form - so it would clear the foreign key relationship.
		Select(fieldSetWithUpdatedAt(relationModelStruct, relField.Relationship().ForeignKey())...).
		// Where all relation models where the foreign key is one of the roots primaries.
		Filter(filter.New(relField.Relationship().ForeignKey(), filter.OpIn, primaryKeyValues...)).
		Update()
}

func clearHasManyRelations(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField) (int64, error) {
	// Removing relations from the HasOne relationships involves setting related foreign key value to zero.
	if !relField.Relationship().ForeignKey().IsPtr() {
		return 0, errors.Wrapf(mapping.ErrFieldNotNullable, "provided relation field: '%s' doesn't allow setting null values", relField)
	}
	// If the relationship structure implements update hooks we need to firstly get he
	// related model. The user might set something in the hooks.
	// Otherwise we can change it by using query update.
	relationModel := mapping.NewModel(relField.Relationship().RelatedModelStruct())
	// Check if the related model implements update hooks.
	_, gettingRelationRequired := relationModel.(BeforeUpdater)
	if !gettingRelationRequired {
		_, gettingRelationRequired = relationModel.(AfterUpdater)
	}

	// Get all relationship models where foreign key value is equal to 'models' primary key.
	var primaryKeyValues []interface{}
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.Wrapf(query.ErrInvalidModels, "one of the model values has zero primary key value")
		}
		primaryKeyValues = append(primaryKeyValues, model.GetPrimaryKeyHashableValue())
	}
	relationModelStruct := relField.Relationship().RelatedModelStruct()
	if gettingRelationRequired {
		relationModels, err := db.QueryCtx(ctx, relationModelStruct).
			Select(relationModelStruct.Primary()).
			Filter(filter.New(relField.Relationship().ForeignKey(), filter.OpIn, primaryKeyValues...)).
			Find()
		if err != nil {
			return 0, err
		}

		// update relationship models using their primary and foreign key fields.
		// The foreign key has it's own zero value - previous query pulled only
		return db.QueryCtx(ctx, relationModelStruct, relationModels...).
			Select(fieldSetWithUpdatedAt(relationModelStruct, relationModelStruct.Primary(), relField.Relationship().ForeignKey())...).
			Update()
	}
	// update using queryUpdate - set foreign key values to zero.
	return db.QueryCtx(ctx, relationModelStruct, mapping.NewModel(relationModelStruct)).
		// Select a foreign key - it must be in a zero like form - so it would clear the foreign key relationship.
		Select(fieldSetWithUpdatedAt(relationModelStruct, relField.Relationship().ForeignKey())...).
		// Where all relation models where the foreign key is one of the roots primaries.
		Filter(filter.New(relField.Relationship().ForeignKey(), filter.OpIn, primaryKeyValues...)).
		Update()
}

func clearMany2ManyRelations(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField) (int64, error) {
	// If the relationship structure implements update hooks we need to firstly get he
	// related model. The user might set something in the hooks.
	// Otherwise we can change it by using query update.
	// Check if the related model implements update hooks.
	joinModelStruct := relField.Relationship().JoinModel()
	joinModel := mapping.NewModel(joinModelStruct)
	_, gettingRelationRequired := joinModel.(BeforeDeleter)
	if !gettingRelationRequired {
		_, gettingRelationRequired = joinModel.(AfterDeleter)
	}

	// Get all relationship models where foreign key value is equal to 'models' primary key.
	var primaryKeyValues []interface{}
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.Wrapf(query.ErrInvalidModels, "one of the model values has zero primary key value")
		}
		primaryKeyValues = append(primaryKeyValues, model.GetPrimaryKeyHashableValue())
	}
	if gettingRelationRequired {
		joinModels, err := db.QueryCtx(ctx, joinModelStruct).
			Select(joinModelStruct.Primary()).
			Filter(filter.New(relField.Relationship().ForeignKey(), filter.OpIn, primaryKeyValues...)).
			Find()
		if err != nil {
			return 0, err
		}

		// deleteQuery join models using their primary and foreign key fields.
		// The foreign key has it's own zero value - previous query pulled only
		return db.QueryCtx(ctx, joinModelStruct, joinModels...).Delete()
	}
	// deleteQuery all set foreign key values to zero.
	return db.QueryCtx(ctx, joinModelStruct, mapping.NewModel(joinModelStruct)).
		// Where all relation models where the foreign key is one of the roots primaries.
		Filter(filter.New(relField.Relationship().ForeignKey(), filter.OpIn, primaryKeyValues...)).
		Delete()
}

//
// addModel relations
//

func querySetRelations(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels ...mapping.Model) error {
	if !relationField.IsRelationship() {
		return errors.Wrapf(query.ErrInvalidField, "provided field is not a relation: '%s' in model: '%s'", relationField, s.ModelStruct)
	}
	if len(s.Models) == 0 {
		return errors.Wrap(query.ErrInvalidInput, "no models provided for the query")
	}
	if len(relationModels) == 0 {
		return errors.Wrap(query.ErrInvalidInput, "no relation models provided for the query")
	}

	switch relationField.Relationship().Kind() {
	case mapping.RelBelongsTo:
		return setBelongsToRelation(ctx, db, s, relationField, relationModels)
	case mapping.RelHasOne:
		return setHasOneRelation(ctx, db, s, relationField, relationModels)
	case mapping.RelHasMany:
		return setHasManyRelations(ctx, db, s, relationField, relationModels)
	case mapping.RelMany2Many:
		return setMany2ManyRelations(ctx, db, s, relationField, relationModels)
	default:
		return errors.Wrapf(query.ErrInternal, "invalid relationship kind: '%s'", relationField.Relationship().Kind())
	}
}

func setMany2ManyRelations(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) (err error) {
	if _, err = clearMany2ManyRelations(ctx, db, s, relationField); err != nil {
		return err
	}
	return addRelationMany2Many(ctx, db, s, relationField, relationModels)
}

func setHasOneRelation(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) (err error) {
	if len(s.Models) > 1 {
		return errors.Wrap(query.ErrInvalidInput, "cannot set many relations to multiple models with hasMany relationship")
	}
	if len(relationModels) > 1 {
		return errors.Wrap(query.ErrInvalidInput, "cannot set multiple relations for single has one model")
	}
	if _, err = clearHasOneRelation(ctx, db, s, relationField); err != nil {
		return err
	}
	return addRelationHasOne(ctx, db, s, relationField, relationModels)
}

func setHasManyRelations(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) (err error) {
	if len(s.Models) > 1 {
		return errors.Wrap(query.ErrInvalidInput, "cannot set many relations to multiple models with hasMany relationship")
	}
	if _, err = clearHasManyRelations(ctx, db, s, relationField); err != nil {
		return err
	}
	return addRelationHasMany(ctx, db, s, relationField, relationModels)
}

func setBelongsToRelation(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) error {
	if len(relationModels) > 1 {
		return errors.Wrap(query.ErrInvalidInput, "cannot set multiple relation models to the models with belongs to relationship")
	}
	if relationModels[0].IsPrimaryKeyZero() {
		return errors.Wrapf(query.ErrInvalidInput, "provided relation has zero value primary key")
	}
	relationPrimary := relationModels[0].GetPrimaryKeyHashableValue()

	// If the model has UpdatedAt timestamp set it manually.
	updatedAt, hasUpdated := s.ModelStruct.UpdatedAt()
	tsNow := db.Controller().Now()

	// For models in the query scope set foreign key field value to the relation's primary.
	for i, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return errors.Wrapf(query.ErrInvalidModels, "model[%d] has zero value primary key", i)
		}
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return errModelNotImplements(s.ModelStruct, "Fielder")
		}
		if err := fielder.SetFieldValue(relationField.Relationship().ForeignKey(), relationPrimary); err != nil {
			return err
		}
		if hasUpdated {
			if err := fielder.SetFieldValue(updatedAt, tsNow); err != nil {
				return err
			}
		}
	}

	// The fieldset should contain only
	s.FieldSets = append(s.FieldSets, mapping.FieldSet{s.ModelStruct.Primary(), relationField.Relationship().ForeignKey()})
	if hasUpdated {
		s.FieldSets[0] = append(s.FieldSets[0], updatedAt)
	}

	// update models with new foreign key field values.
	_, err := queryUpdate(ctx, db, s)
	return err
}

// queryIncludeRelation gets related model for the 'models' used in a relation from field 'relationField'.
// An optional relationFieldset could be used in order to specify fields for the relations.
func queryIncludeRelation(ctx context.Context, db DB, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) error {
	s := query.NewScope(mStruct, models...)

	if err := s.Include(relationField, relationFieldset...); err != nil {
		return err
	}
	if err := findIncludedRelationsSynchronous(ctx, db, s); err != nil {
		return err
	}
	return nil
}

// queryFindIncludedRelations find included relations for all models stored in given query and for all included relation fields.
func queryFindIncludedRelations(ctx context.Context, db DB, s *query.Scope) error {
	if len(s.Models) == 0 {
		return errors.WrapDetf(query.ErrNoModels, "provided no models in the query")
	}
	if len(s.IncludedRelations) == 0 {
		return errors.WrapDetf(query.ErrInvalidInput, "provided no included relations")
	}
	return findIncludedRelations(ctx, db, s)
}

// queryGetRelations gets the relations - at 'relationField' with optional 'relationFieldset' for provided 'models'.
func queryGetRelations(ctx context.Context, db DB, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error) {
	if len(models) == 0 {
		return nil, errors.WrapDetf(query.ErrNoModels, "no input models provided")
	}
	if !relationField.IsRelationship() {
		return nil, errors.WrapDetf(mapping.ErrInvalidRelationField, "provided field is not a relationship")
	}
	switch relationField.Relationship().Kind() {
	case mapping.RelBelongsTo:
		return getRelationsBelongsTo(ctx, db, mStruct, models, relationField, relationFieldset...)
	case mapping.RelHasMany, mapping.RelHasOne:
		return getRelationsHasOneOrMany(ctx, db, mStruct, models, relationField, relationFieldset...)
	case mapping.RelMany2Many:
		return getRelationsMany2Many(ctx, db, mStruct, models, relationField, relationFieldset...)
	default:
		return nil, errors.Wrapf(query.ErrInternal, "invalid relationship kind: '%s'", relationField.Relationship().Kind())
	}
}

func getRelationsBelongsTo(ctx context.Context, db DB, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error) {
	relationship := relationField.Relationship()
	if len(relationFieldset) == 0 {
		relationFieldset = relationship.RelatedModelStruct().Fields()
	}
	foreignKeys := map[interface{}]struct{}{}
	for _, model := range models {
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return nil, errors.WrapDetf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder interface", mStruct.String())
		}
		// Check if the foreign key value is not 'zero'.
		isZero, err := fielder.IsFieldZero(relationship.ForeignKey())
		if err != nil {
			return nil, err
		}
		if isZero {
			continue
		}

		foreignKeyValue, err := fielder.GetHashableFieldValue(relationship.ForeignKey())
		if err != nil {
			return nil, err
		}
		foreignKeys[foreignKeyValue] = struct{}{}
	}
	if len(foreignKeys) == 0 {
		return []mapping.Model{}, nil
	}

	// Get unique foreign key filter values.
	filterValues := make([]interface{}, len(foreignKeys))
	for k := range foreignKeys {
		filterValues = append(filterValues, k)
	}
	relatedModels, err := db.QueryCtx(ctx, relationship.RelatedModelStruct()).
		Filter(filter.New(relationship.RelatedModelStruct().Primary(), filter.OpIn, filterValues...)).
		Select(relationFieldset...).
		Find()
	if err != nil {
		return nil, err
	}
	return relatedModels, nil
}

func getRelationsHasOneOrMany(ctx context.Context, db DB, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error) {
	relationship := relationField.Relationship()
	if len(relationFieldset) == 0 {
		relationFieldset = relationship.RelatedModelStruct().Fields()
	}
	var filterValues []interface{}
	// Iterate over scope models and map their primary key values to their indexes in the scope value slice.
	for _, model := range models {
		if model.IsPrimaryKeyZero() {
			continue
		}

		filterValues = append(filterValues, model.GetPrimaryKeyValue())
	}

	relatedModels, err := db.QueryCtx(ctx, relationship.RelatedModelStruct()).
		Filter(filter.New(relationship.ForeignKey(), filter.OpIn, filterValues...)).
		Select(relationFieldset...).
		Find()
	if err != nil {
		return nil, err
	}
	return relatedModels, nil
}

func getRelationsMany2Many(ctx context.Context, db DB, mStruct *mapping.ModelStruct, models []mapping.Model, relationField *mapping.StructField, relationFieldset ...*mapping.StructField) ([]mapping.Model, error) {
	relationship := relationField.Relationship()
	if len(relationFieldset) == 0 {
		relationFieldset = relationship.RelatedModelStruct().Fields()
	}

	// Map the primary key field value to the slice index. Store the values in the
	var filterValues []interface{}
	for _, model := range models {
		if model.IsPrimaryKeyZero() {
			continue
		}
		filterValues = append(filterValues, model.GetPrimaryKeyValue())
	}

	// Get the back reference foreign key - a field that relates to the root models primary key field in the join table.
	backReferenceFK := relationship.ForeignKey()
	// Get the join model foreign key - a field that relates to the relation model primary key in the join table.
	joinModelFK := relationship.ManyToManyForeignKey()

	joinModels, err := db.QueryCtx(ctx, relationship.JoinModel()).
		Select(backReferenceFK, joinModelFK).
		Filter(filter.New(backReferenceFK, filter.OpIn, filterValues...)).
		Find()
	if err != nil {
		return nil, err
	}
	if len(joinModels) == 0 {
		return []mapping.Model{}, nil
	}

	joinModelForeignKeys := map[interface{}]struct{}{}
	// If the field was included the relations needs to find with new scope.
	// Iterate over relatedScope Models
	for _, model := range joinModels {
		if model == nil {
			continue
		}
		fielder, ok := model.(mapping.Fielder)
		if !ok {
			return nil, errors.Wrapf(mapping.ErrModelNotImplements, "model: '%s' doesn't implement Fielder interface", model.NeuronCollectionName())
		}
		isZero, err := fielder.IsFieldZero(backReferenceFK)
		if err != nil {
			return nil, err
		}
		if isZero {
			continue
		}
		foreignKeyValue, err := fielder.GetHashableFieldValue(backReferenceFK)
		if err != nil {
			return nil, err
		}
		joinModelForeignKeys[foreignKeyValue] = struct{}{}
	}
	if len(joinModelForeignKeys) == 0 {
		return []mapping.Model{}, nil
	}

	if len(relationFieldset) == 1 && relationFieldset[0] == relationship.RelatedModelStruct().Primary() {
		result := make([]mapping.Model, len(joinModelForeignKeys))
		var i int
		for fk := range joinModelForeignKeys {
			model := mapping.NewModel(relationship.RelatedModelStruct())
			if err := model.SetPrimaryKeyValue(fk); err != nil {
				return nil, err
			}
			result[i] = model
			i++
		}
		return result, nil
	}

	// create a scope for the relation
	var relationPrimaries []interface{}
	for k := range joinModelForeignKeys {
		relationPrimaries = append(relationPrimaries, k)
	}

	relationModels, err := db.QueryCtx(ctx, relationship.RelatedModelStruct()).
		Filter(filter.New(relationship.RelatedModelStruct().Primary(), filter.OpIn, relationPrimaries...)).
		Select(relationFieldset...).
		Find()
	if err != nil {
		return []mapping.Model{}, err
	}
	return relationModels, nil
}
