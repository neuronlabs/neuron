package scope

import (
	"github.com/neuronlabs/neuron-core/log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/namer"

	"github.com/neuronlabs/neuron-core/internal/models"
)

// User is the include testing model.
type User struct {
	ID   int
	Cars []*Car
}

// Car is the included testing model.
type Car struct {
	ID     int
	UserID int

	Brand   *CarBrand
	BrandID int `neuron:"type=fk;name=brand_id"`
}

// CarBrand is the included testing model.
type CarBrand struct {
	ID int
}

// TestBuildIncludedFields tests the BuildIncludedFields function.
func TestBuildIncludedFields(t *testing.T) {
	if testing.Verbose() {
		log.Default()
		log.SetLevel(log.LDEBUG3)
	}
	t.Run("Valid", func(t *testing.T) {
		m := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

		err := m.RegisterModels(User{}, Car{}, CarBrand{})
		require.NoError(t, err)

		userModel := m.GetByCollection("users")
		require.NotNil(t, userModel)

		s := newScope(userModel)

		errs := s.BuildIncludedFields("cars", "cars.brand")
		assert.Empty(t, errs)
	})

	t.Run("InvalidFieldName", func(t *testing.T) {
		m := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

		err := m.RegisterModels(User{}, Car{}, CarBrand{})
		require.NoError(t, err)

		userModel := m.GetByCollection("users")
		require.NotNil(t, userModel)

		s := newScope(userModel)

		errs := s.BuildIncludedFields("cares")
		assert.NotEmpty(t, errs)
	})

	t.Run("Duplicated", func(t *testing.T) {
		m := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

		err := m.RegisterModels(User{}, Car{}, CarBrand{})
		require.NoError(t, err)

		userModel := m.GetByCollection("users")
		require.NotNil(t, userModel)

		s := newScope(userModel)

		errs := s.BuildIncludedFields("cars", "cars")
		assert.NotEmpty(t, errs)
	})

	t.Run("NestedDuplicated", func(t *testing.T) {
		m := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

		err := m.RegisterModels(User{}, Car{}, CarBrand{})
		require.NoError(t, err)

		userModel := m.GetByCollection("users")
		require.NotNil(t, userModel)

		s := newScope(userModel)

		errs := s.BuildIncludedFields("cars", "cars.brand", "cars.brand")
		assert.NotEmpty(t, errs)
	})

	t.Run("TooManyIncludes", func(t *testing.T) {
		cfg := config.ReadDefaultControllerConfig()
		cfg.IncludedDepthLimit = 0
		m := models.NewModelMap(namer.NamingSnake, cfg)

		err := m.RegisterModels(User{}, Car{}, CarBrand{})
		require.NoError(t, err)

		userModel := m.GetByCollection("users")
		require.NotNil(t, userModel)

		s := newScope(userModel)

		errs := s.BuildIncludedFields("cars.brand")
		assert.NotEmpty(t, errs, "%v", errs)
	})

	t.Run("MaxDepthReached", func(t *testing.T) {
		cfg := config.ReadDefaultControllerConfig()
		cfg.IncludedDepthLimit = 0
		m := models.NewModelMap(namer.NamingSnake, cfg)

		err := m.RegisterModels(User{}, Car{}, CarBrand{})
		require.NoError(t, err)

		userModel := m.GetByCollection("users")
		require.NotNil(t, userModel)

		s := newScope(userModel)

		errs := s.BuildIncludedFields("cars", "cars.brand")
		assert.NotEmpty(t, errs, "%v", errs)
	})
}

// TestIncludeGetMissingPrimaries test GetMissingPrimaries method for the included fields.
func TestIncludeGetMissingPrimaries(t *testing.T) {
	t.Run("HasMany", func(t *testing.T) {
		m := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

		err := m.RegisterModels(User{}, Car{}, CarBrand{})
		require.NoError(t, err)

		userModel := m.GetByCollection("users")
		require.NotNil(t, userModel)

		s := newScope(userModel)

		// Include the cars field
		errs := s.BuildIncludedFields("cars")
		require.Empty(t, errs)

		usersValue := []*User{{ID: 1, Cars: []*Car{{ID: 1}, {ID: 2}}}, {ID: 2, Cars: []*Car{{ID: 2}, {ID: 4}}}}
		s.Value = &usersValue

		includedFields := s.IncludedFields()
		require.Len(t, includedFields, 1)

		carsField := includedFields[0]
		assert.Equal(t, "cars", carsField.StructField.NeuronName())

		primaries, err := carsField.GetMissingPrimaries()
		if assert.NoError(t, err) {
			if assert.Len(t, primaries, 3) {
				assert.Contains(t, primaries, 1)
				assert.Contains(t, primaries, 2)
				assert.Contains(t, primaries, 4)
			}
		}
	})

	t.Run("BelongsTo", func(t *testing.T) {
		m := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

		err := m.RegisterModels(User{}, Car{}, CarBrand{})
		require.NoError(t, err)

		carModel := m.GetByCollection("cars")
		require.NotNil(t, carModel)

		s := newScope(carModel)

		// Include the cars field
		errs := s.BuildIncludedFields("brand")
		require.Empty(t, errs)

		carsValue := []*Car{{ID: 1, Brand: &CarBrand{ID: 16}}, {ID: 2, Brand: &CarBrand{ID: 21}}}
		s.Value = &carsValue

		includedFields := s.IncludedFields()
		require.Len(t, includedFields, 1)

		s.NextIncludedField()

		brandField, err := s.CurrentIncludedField()
		require.NoError(t, err)

		assert.Equal(t, "brand", brandField.StructField.NeuronName())

		primaries, err := brandField.GetMissingPrimaries()
		if assert.NoError(t, err) {
			if assert.Len(t, primaries, 2) {
				assert.Contains(t, primaries, 16)
				assert.Contains(t, primaries, 21)
			}
		}

		s.NextIncludedField()
		_, err = s.CurrentIncludedField()
		require.Error(t, err)
	})
}

// TestIncludeCopyScopeCommon tests the common included field methods.
func TestIncludeCopyScopeCommon(t *testing.T) {
	m := models.NewModelMap(namer.NamingSnake, config.ReadDefaultControllerConfig())

	err := m.RegisterModels(User{}, Car{}, CarBrand{})
	require.NoError(t, err)

	userModel := m.GetByCollection("users")
	require.NotNil(t, userModel)

	s := newScope(userModel)

	// Include the cars field
	errs := s.BuildIncludedFields("cars")
	require.Empty(t, errs)

	usersValue := []*User{{ID: 1, Cars: []*Car{{ID: 1}, {ID: 2}}}, {ID: 2, Cars: []*Car{{ID: 2}, {ID: 4}}}}
	s.Value = &usersValue

	includedFields := s.IncludedFields()
	require.Len(t, includedFields, 1)

	carsField := includedFields[0]
	assert.Equal(t, "cars", carsField.StructField.NeuronName())

	includedScopes := s.IncludedScopes()
	if assert.Len(t, includedScopes, 1) {
		includedScope := includedScopes[0]
		scope, ok := s.IncludedScopeByStruct(carsField.StructField.Relationship().Struct())
		if assert.True(t, ok) {
			assert.Equal(t, includedScope, scope)
		}
	}
	s.CopyIncludedBoundaries()

	var fromChan []*IncludeField
	channel := s.IncludedFieldsChan()
	for field := range channel {
		fromChan = append(fromChan, field)
	}

	assert.Equal(t, len(fromChan), len(s.IncludedFields()))
}
