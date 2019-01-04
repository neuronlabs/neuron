package controller

import (
	"fmt"
	"github.com/kucjac/jsonapi/pkg/mapping"
	"github.com/pkg/errors"

	"reflect"
	"sync"
)

var (
	errBadJSONAPIStructTag = errors.New("Bad jsonapi struct tag format")
	IErrModelNotMapped     = errors.New("Unmapped model provided.")
)

// ModelMap contains mapped models ( as reflect.Type ) to its ModelStruct representation.
// Allow concurrent safe gets and sets to the map.
type ModelMap struct {
	models      map[reflect.Type]*mapping.ModelStruct
	collections map[string]reflect.Type
	sync.RWMutex
}

func newModelMap() *ModelMap {
	var modelMap *ModelMap = &ModelMap{
		models:      make(map[reflect.Type]*mapping.ModelStruct),
		collections: make(map[string]reflect.Type),
	}

	return modelMap
}

// Set is concurrent safe setter for model structs
func (m *ModelMap) Set(key reflect.Type, value *mapping.ModelStruct) {
	m.Lock()
	defer m.Unlock()
	m.models[key] = value
}

// Get is concurrent safe getter of model structs.
func (m *ModelMap) Get(key reflect.Type) *mapping.ModelStruct {
	m.RLock()
	defer m.RUnlock()
	return m.models[key]
}

func (m *ModelMap) GetByCollection(collection string) *mapping.ModelStruct {
	m.RLock()
	defer m.RUnlock()
	t, ok := m.collections[collection]
	if !ok || t == nil {
		return nil
	}
	return m.models[t]
}

func (m *ModelMap) getSimilarCollections(collection string) (simillar []string) {
	/**

	TO IMPLEMENT:

	find closest match collection

	*/
	return []string{}
}

func getSliceElemType(modelType reflect.Type) (reflect.Type, error) {
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Slice {
		err := fmt.Errorf("Invalid input for slice elem type: %v", modelType)
		return modelType, err
	}

	for modelType.Kind() == reflect.Ptr || modelType.Kind() == reflect.Slice {
		modelType = modelType.Elem()
	}

	return modelType, nil
}
