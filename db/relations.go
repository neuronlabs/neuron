package db

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// AddRelations appends relationship 'relField' 'relationModels' to the scope values.
func AddRelations(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField, relationModels ...mapping.Model) error {
	if s.ModelStruct != relField.Struct() {
		return errors.Newf(query.ClassInvalidField, "provided relation field: '%s' doesn't belong to the scope's relationModel: '%s'", relField.String(), s.ModelStruct.String())
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
		return errors.Newf(query.ClassInvalidField, "provided relation field: '%s' with invalid relationship kind: '%s'", relField.String(), relField.Relationship().Kind().String())
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
			return errors.New(query.ClassInvalidModels, "one of provided model's primary key value has zero value")
		}

		// For each related model insert join model instance with the foreign key value from 'model' primary key
		// and many2many foreign key value from relationModel primary key.
		for _, relationModel := range relationModels {
			if relationModel.IsPrimaryKeyZero() {
				return errors.New(query.ClassInvalidModels, "one of provided relation model's primary key value has zero value")
			}

			joinModel := mapping.NewModel(relationship.JoinModel())
			joinModelFielder, ok := joinModel.(mapping.Fielder)
			if !ok {
				return errors.Newf(mapping.ClassModelNotImplements, "model: '%s' doesn't implement Fielder interface", joinModel.NeuronCollectionName())
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
		return errors.Newf(query.ClassInvalidModels, "relation field: '%s' of HasMany kind cannot be added to multiple model values", relField.String())
	}
	// Iterate over relationship models and set relationship foreign key value to the
	// scope's primary key value.
	model := s.Models[0]
	if model == nil {
		return errors.Newf(query.ClassInvalidModels, "provided nil model value")
	}
	// Check if the model's primary key is non zero.
	if model.IsPrimaryKeyZero() {
		return errors.New(query.ClassInvalidInput, "provided model's primary key value has zero value")
	}
	relationship := relField.Relationship()
	updatedAt, hasUpdatedAt := relationship.Struct().UpdatedAt()
	updatedAtTS := db.Controller().Now()
	// Iterate over all relation models and set relation's foreign key to the primary key of the 'model'.
	for _, relationModel := range relationModels {
		if relationModel.IsPrimaryKeyZero() {
			return errors.Newf(query.ClassInvalidInput, "relation model value has zero primary key value")
		}
		// addModel the foreign key field to the model's primary key value.
		fielder, ok := relationModel.(mapping.Fielder)
		if !ok {
			return errModelNotImplements(relationship.Struct(), "Fielder")
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
	// Update all relation models using their primary key as well as this relationship foreign key.
	_, err := db.
		QueryCtx(ctx, relationship.Struct(), relationModels...).
		Select(fieldSetWithUpdatedAt(relationship.Struct(), relationship.Struct().Primary(), relationship.ForeignKey())...).
		Update()
	return err
}

func addRelationHasOne(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) error {
	if len(relationModels) > 1 {
		return errors.New(query.ClassInvalidInput, "cannot set multiple relation models to the models with 'has one' relationship")
	}
	relationModel := relationModels[0]
	if relationModel.IsPrimaryKeyZero() {
		return errors.New(query.ClassInvalidInput, "provided relation has zero value primary key")
	}
	if len(s.Models) > 1 {
		return errors.New(query.ClassInvalidInput, "cannot set multiple models with 'has one' relation ship to the single relation field")
	}

	model := s.Models[0]
	if model.IsPrimaryKeyZero() {
		return errors.New(query.ClassInvalidInput, "provide model has zero value primary key")
	}

	relationship := relationField.Relationship()

	fielder, ok := relationModel.(mapping.Fielder)
	if !ok {
		return errModelNotImplements(relationship.Struct(), "Fielder")
	}
	if err := fielder.SetFieldValue(relationship.ForeignKey(), model.GetPrimaryKeyHashableValue()); err != nil {
		return err
	}

	fieldSet := mapping.FieldSet{relationship.Struct().Primary(), relationship.ForeignKey()}

	updatedAt, hasUpdatedAt := relationship.Struct().UpdatedAt()
	if hasUpdatedAt {
		if err := fielder.SetFieldValue(updatedAt, db.Controller().Now()); err != nil {
			return err
		}
		fieldSet = append(fieldSet, updatedAt)
	}

	_, err := db.QueryCtx(ctx, relationship.Struct(), relationModel).
		Select(fieldSet...).
		Update()
	return err
}

//
// Remove relations.
//

func RemoveRelations(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField) (int64, error) {
	if s.ModelStruct != relField.Struct() {
		return 0, errors.Newf(query.ClassInvalidField, "provided relation field: '%s' doesn't belong to the scope's relationModel: '%s'", relField.String(), s.ModelStruct.String())
	}
	if !relField.IsRelationship() {
		return 0, errors.Newf(query.ClassInvalidField, "provided field: '%s' is not a relationship", relField)
	}
	if err := requireNoFilters(s); err != nil {
		return 0, err
	}
	if len(s.Models) == 0 {
		return 0, errors.Newf(query.ClassNoModels, "provided no models to remove the relations")
	}
	switch relField.Relationship().Kind() {
	case mapping.RelMany2Many:
		return removeMany2ManyRelations(ctx, db, s, relField)
	case mapping.RelHasOne, mapping.RelHasMany:
		return removeHasRelations(ctx, db, s, relField)
	case mapping.RelBelongsTo:
		return removeBelongsToRelation(ctx, db, s, relField)
	default:
		return 0, errors.Newf(query.ClassInternal, "invalid relationship kind: '%s'", relField.Relationship().Kind())
	}
}

func removeBelongsToRelation(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField) (int64, error) {
	// Removing relations from the belongs to relationship involves setting foreign key field's value to nullable.
	if !relField.Relationship().ForeignKey().IsPtr() {
		return 0, errors.Newf(query.ClassInvalidField, "provided relation field: '%s' doesn't allow setting null values", relField)
	}
	if err := requireNoFilters(s); err != nil {
		return 0, err
	}
	// Iterate over scope models and set relationship foreign key value to zero - nullable.
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.Newf(query.ClassInvalidModels, "one of the model values has zero primary key value")
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
	return Update(ctx, db, s)
}

func removeHasRelations(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField) (int64, error) {
	// Removing relations from the HasOne relationships involves setting related foreign key value to zero.
	if !relField.Relationship().ForeignKey().IsPtr() {
		return 0, errors.Newf(query.ClassInvalidField, "provided relation field: '%s' doesn't allow setting null values", relField)
	}
	// If the relationship structure implements Update hooks we need to firstly get he
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
			return 0, errors.Newf(query.ClassInvalidModels, "one of the model values has zero primary key value")
		}
		primaryKeyValues = append(primaryKeyValues, model.GetPrimaryKeyHashableValue())
	}
	relationModelStruct := relField.Relationship().Struct()
	if gettingRelationRequired {
		relationModels, err := db.QueryCtx(ctx, relationModelStruct).
			Select(relationModelStruct.Primary()).
			Filter(filter.New(relField.Relationship().ForeignKey(), filter.OpIn, primaryKeyValues...)).
			Find()
		if err != nil {
			return 0, err
		}

		// Update relationship models using their primary and foreign key fields.
		// The foreign key has it's own zero value - previous query pulled only
		return db.QueryCtx(ctx, relationModelStruct, relationModels...).
			Select(fieldSetWithUpdatedAt(relationModelStruct, relationModelStruct.Primary(), relField.Relationship().ForeignKey())...).
			Update()
	}
	// Update using queryUpdate - set foreign key values to zero.
	return db.QueryCtx(ctx, relationModelStruct, mapping.NewModel(relationModelStruct)).
		// Select a foreign key - it must be in a zero like form - so it would clear the foreign key relationship.
		Select(fieldSetWithUpdatedAt(relationModelStruct, relField.Relationship().ForeignKey())...).
		// Where all relation models where the foreign key is one of the roots primaries.
		Filter(filter.New(relField.Relationship().ForeignKey(), filter.OpIn, primaryKeyValues...)).
		Update()
}

func removeMany2ManyRelations(ctx context.Context, db DB, s *query.Scope, relField *mapping.StructField) (int64, error) {
	// Removing relations from the HasOne relationships involves setting related foreign key value to zero.
	if !relField.Relationship().ForeignKey().IsPtr() {
		return 0, errors.Newf(query.ClassInvalidField, "provided relation field: '%s' doesn't allow setting null values", relField)
	}
	// If the relationship structure implements Update hooks we need to firstly get he
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
	primaryKeyValues := []interface{}{}
	for _, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return 0, errors.Newf(query.ClassInvalidModels, "one of the model values has zero primary key value")
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

		// Delete join models using their primary and foreign key fields.
		// The foreign key has it's own zero value - previous query pulled only
		return db.QueryCtx(ctx, joinModelStruct, joinModels...).Delete()
	}
	// Delete all set foreign key values to zero.
	return db.QueryCtx(ctx, joinModelStruct, mapping.NewModel(joinModelStruct)).
		// Where all relation models where the foreign key is one of the roots primaries.
		Filter(filter.New(relField.Relationship().ForeignKey(), filter.OpIn, primaryKeyValues...)).
		Delete()
}

//
// addModel relations
//

func SetRelations(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels ...mapping.Model) error {
	if !relationField.IsRelationship() {
		return errors.Newf(query.ClassInvalidField, "provided field is not a relation: '%s' in model: '%s'", relationField, s.ModelStruct)
	}
	if len(s.Models) == 0 {
		return errors.New(query.ClassInvalidInput, "no models provided for the query")
	}
	if len(relationModels) == 0 {
		return errors.New(query.ClassInvalidInput, "no relation models provided for the query")
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
		return errors.Newf(query.ClassInternal, "invalid relationship kind: '%s'", relationField.Relationship().Kind())
	}
}

func setMany2ManyRelations(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) (err error) {
	if _, err = removeMany2ManyRelations(ctx, db, s, relationField); err != nil {
		return err
	}
	return addRelationMany2Many(ctx, db, s, relationField, relationModels)
}

func setHasOneRelation(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) (err error) {
	if len(s.Models) > 1 {
		return errors.New(query.ClassInvalidInput, "cannot set many relations to multiple models with hasMany relationship")
	}
	if len(relationModels) > 1 {
		return errors.New(query.ClassInvalidInput, "cannot set multiple relations for single has one model")
	}
	if _, err = removeHasRelations(ctx, db, s, relationField); err != nil {
		return err
	}
	return addRelationHasOne(ctx, db, s, relationField, relationModels)
}

func setHasManyRelations(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) (err error) {
	if len(s.Models) > 1 {
		return errors.New(query.ClassInvalidInput, "cannot set many relations to multiple models with hasMany relationship")
	}
	if _, err = removeHasRelations(ctx, db, s, relationField); err != nil {
		return err
	}
	return addRelationHasMany(ctx, db, s, relationField, relationModels)
}

func setBelongsToRelation(ctx context.Context, db DB, s *query.Scope, relationField *mapping.StructField, relationModels []mapping.Model) error {
	if len(relationModels) > 1 {
		return errors.New(query.ClassInvalidInput, "cannot set multiple relation models to the models with belongs to relationship")
	}
	if relationModels[0].IsPrimaryKeyZero() {
		return errors.Newf(query.ClassInvalidInput, "provided relation has zero value primary key")
	}
	relationPrimary := relationModels[0].GetPrimaryKeyHashableValue()

	// If the model has UpdatedAt timestamp set it manually.
	updatedAt, hasUpdated := s.ModelStruct.UpdatedAt()
	tsNow := db.Controller().Now()

	// For models in the query scope set foreign key field value to the relation's primary.
	for i, model := range s.Models {
		if model.IsPrimaryKeyZero() {
			return errors.Newf(query.ClassInvalidModels, "model[%d] has zero value primary key", i)
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

	// Update models with new foreign key field values.
	_, err := Update(ctx, db, s)
	return err
}
