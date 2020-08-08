package database

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
)

// refreshQuery refreshed models from the provided query.
func refreshQuery(ctx context.Context, db DB, q *query.Scope) error {
	switch len(q.Models) {
	case 0:
		return errors.WrapDet(query.ErrInvalidInput, "nothing to refresh")
	case 1:
		return refreshSingle(ctx, db, q)
	default:
		return refreshMany(ctx, db, q)
	}
}

func refreshSingle(ctx context.Context, db DB, q *query.Scope) error {
	model := q.Models[0]
	if model.IsPrimaryKeyZero() {
		return errors.WrapDetf(query.ErrInvalidInput, "provided model has zero value primary key")
	}
	q.Models = nil
	q.Filter(filter.New(q.ModelStruct.Primary(), filter.OpEqual, model.GetPrimaryKeyValue()))
	result, err := queryGet(ctx, db, q)
	if err != nil {
		return err
	}
	if err = setFrom(q.ModelStruct, model, result); err != nil {
		return err
	}
	return nil
}

func refreshMany(ctx context.Context, db DB, q *query.Scope) error {
	models := q.Models
	idToIndex := map[interface{}]int{}
	primaryKeys := make([]interface{}, len(models))
	for i, model := range models {
		if model.IsPrimaryKeyZero() {
			return errors.WrapDetf(query.ErrInvalidInput, "one of provided models has zero value primary key")
		}
		idToIndex[model.GetPrimaryKeyHashableValue()] = i
		primaryKeys[i] = model.GetPrimaryKeyValue()
	}

	// Create new primary key filter and copy each field from the results to the input models.
	q.Models = nil
	q.Filter(filter.New(q.ModelStruct.Primary(), filter.OpIn, primaryKeys...))
	results, err := queryFind(ctx, db, q)
	if err != nil {
		return err
	}

	for _, result := range results {
		index, ok := idToIndex[result.GetPrimaryKeyHashableValue()]
		if !ok {
			continue
		}
		model := models[index]
		if err := setFrom(q.ModelStruct, model, result); err != nil {
			return err
		}
	}
	return nil
}

func setFrom(mStruct *mapping.ModelStruct, to, from mapping.Model, includedRelations ...*query.IncludedRelation) error {
	fromSetter, ok := to.(mapping.FromSetter)
	if ok {
		if err := fromSetter.SetFrom(from); err != nil {
			return err
		}
		return nil
	}

	resultFielder, ok := from.(mapping.Fielder)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %s doesn't implement Fielder interface", mStruct)
	}
	modelFielder, ok := to.(mapping.Fielder)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %s doesn't implement Fielder interface", mStruct)
	}
	for _, sField := range mStruct.Fields() {
		fieldValue, err := resultFielder.GetFieldValue(sField)
		if err != nil {
			return err
		}
		if err = modelFielder.SetFieldValue(sField, fieldValue); err != nil {
			return err
		}
	}

	if len(includedRelations) == 0 {
		return nil
	}
	var (
		resultMulti, modelMulti   mapping.MultiRelationer
		resultSingle, modelSingle mapping.SingleRelationer
	)
	for _, included := range includedRelations {
		switch included.StructField.Kind() {
		case mapping.KindRelationshipMultiple:
			if resultMulti == nil {
				resultMulti, ok = from.(mapping.MultiRelationer)
				if !ok {
					return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %s doesn't implement MultiRelationer interface", mStruct)
				}
			}
			if modelMulti == nil {
				modelMulti, ok = to.(mapping.MultiRelationer)
				if !ok {
					return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %s doesn't implement MultiRelationer interface", mStruct)
				}
			}
			relation, err := resultMulti.GetRelationModels(included.StructField)
			if err != nil {
				return err
			}
			if err = modelMulti.SetRelationModels(included.StructField, relation...); err != nil {
				return err
			}
		case mapping.KindRelationshipSingle:
			if resultSingle == nil {
				resultSingle, ok = from.(mapping.SingleRelationer)
				if !ok {
					return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %s doesn't implement MultiRelationer interface", mStruct)
				}
			}
			if modelSingle == nil {
				modelSingle, ok = to.(mapping.SingleRelationer)
				if !ok {
					return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %s doesn't implement MultiRelationer interface", mStruct)
				}
			}
			relation, err := resultSingle.GetRelationModel(included.StructField)
			if err != nil {
				return err
			}
			if err = modelSingle.SetRelationModel(included.StructField, relation); err != nil {
				return err
			}
		default:
			return errors.Wrap(query.ErrInternal, "provided invalid included relation field")
		}
	}
	return nil
}

// RefreshModels refreshes all 'fields' (attributes, foreign keys) for provided input 'models'.
func RefreshModels(ctx context.Context, db DB, mStruct *mapping.ModelStruct, models []mapping.Model, fieldSet ...*mapping.StructField) error {
	idToIndex := map[interface{}]int{}
	primaryKeys := make([]interface{}, len(models))
	for i, model := range models {
		if model.IsPrimaryKeyZero() {
			return errors.WrapDetf(query.ErrInvalidInput, "one of provided models has zero value primary key")
		}
		idToIndex[model.GetPrimaryKeyHashableValue()] = i
		primaryKeys[i] = model.GetPrimaryKeyValue()
	}

	// Create new query scope with the primary key filters and copy each field from the results to the input models.
	q := query.NewScope(mStruct)
	if len(fieldSet) != 0 {
		if err := q.Select(fieldSet...); err != nil {
			return err
		}
	}
	q.Filter(filter.New(mStruct.Primary(), filter.OpIn, primaryKeys...))
	results, err := queryFind(ctx, db, q)
	if err != nil {
		return err
	}

	for _, result := range results {
		index, ok := idToIndex[result.GetPrimaryKeyHashableValue()]
		if !ok {
			continue
		}
		model := models[index]
		resultFielder, ok := result.(mapping.Fielder)
		if !ok {
			return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %s doesn't implement Fielder interface", mStruct)
		}
		modelFielder, ok := model.(mapping.Fielder)
		if !ok {
			return errors.WrapDetf(mapping.ErrModelNotImplements, "model: %s doesn't implement Fielder interface", mStruct)
		}
		for _, sField := range mStruct.Fields() {
			fieldValue, err := resultFielder.GetFieldValue(sField)
			if err != nil {
				return err
			}
			if err = modelFielder.SetFieldValue(sField, fieldValue); err != nil {
				return err
			}
		}
	}
	return nil
}
