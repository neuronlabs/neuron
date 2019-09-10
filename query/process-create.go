package query

import (
	"context"
	"reflect"
	"time"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal"
)

func createFunc(ctx context.Context, s *Scope) error {
	if s.Error != nil {
		return nil
	}

	repo, err := s.Controller().GetRepository(s.Struct())
	if err != nil {
		return err
	}

	creator, ok := repo.(Creator)
	if !ok {
		log.Errorf("The repository doesn't implement Creator interface for model: %s", s.Struct().Collection())
		return errors.NewDet(class.RepositoryNotImplementsCreator, "repository doesn't implement Creator interface")
	}

	// Set created at field if possible
	createdAtField, ok := s.Struct().CreatedAt()
	if ok {
		// by default scope has auto selected fields setCreatedAt should be true
		setCreatedAt := s.autosetFields
		if !setCreatedAt {
			_, found := s.Fieldset[createdAtField.NeuronName()]
			// if the fields were not auto selected check if the field is selected by user
			setCreatedAt = !found
		}

		if setCreatedAt {
			// Check if the value of the created at field is not already set by the user.
			v := reflect.ValueOf(s.Value).Elem().FieldByIndex(createdAtField.ReflectField().Index)

			if s.autosetFields {
				setCreatedAt = reflect.DeepEqual(v.Interface(), reflect.Zero(createdAtField.ReflectField().Type).Interface())
			}

			if setCreatedAt {
				switch {
				case createdAtField.IsTimePointer():
					tv := time.Now()
					v.Set(reflect.ValueOf(&tv))
				case createdAtField.IsTime():
					v.Set(reflect.ValueOf(time.Now()))
				}

				s.Fieldset[createdAtField.NeuronName()] = createdAtField

			}
		}
	}

	if err := creator.Create(ctx, s); err != nil {
		return err
	}

	s.StoreSet(internal.JustCreated, struct{}{})
	return nil
}

// beforeCreate is the function that is used before the create process
func beforeCreateFunc(ctx context.Context, s *Scope) error {
	if s.Error != nil {
		return nil
	}

	beforeCreator, ok := s.Value.(BeforeCreator)
	if !ok {
		return nil
	}

	// Use the hook function before create
	err := beforeCreator.BeforeCreate(ctx, s)
	if err != nil {
		return err
	}
	return nil
}

// afterCreate is the function that is used after the create process
// It uses AfterCreateR hook if the model implements it.
func afterCreateFunc(ctx context.Context, s *Scope) error {
	if s.Error != nil {
		return nil
	}

	afterCreator, ok := s.Value.(AfterCreator)
	if !ok {
		return nil
	}

	err := afterCreator.AfterCreate(ctx, s)
	if err != nil {
		return err
	}
	return nil
}

func storeScopePrimaries(ctx context.Context, s *Scope) error {
	if s.Error != nil {
		return nil
	}

	primaryValues, err := s.getPrimaryFieldValues()
	if err != nil {
		return err
	}

	s.StoreSet(internal.ReducedPrimariesStoreKey, primaryValues)

	return nil
}
