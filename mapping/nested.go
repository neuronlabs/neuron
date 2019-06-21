package mapping

import (
	"reflect"

	"github.com/neuronlabs/neuron/internal/models"
)

// NestedStruct is the structure that represtents nested attribute structure.
type NestedStruct models.NestedStruct

// Attribute returns the attribute on the level of the 'ModelStruct' fields.
func (s *NestedStruct) Attribute() *StructField {
	return (*StructField)((*models.NestedStruct)(s).Attr())
}

// StoreGet gets the value from the store at the key: 'key'.
func (s *NestedStruct) StoreGet(key string) (interface{}, bool) {
	return (*models.NestedStruct)(s).StructField().Self().StoreGet(key)
}

// StoreSet sets into the store the value 'value' for given 'key'.
func (s *NestedStruct) StoreSet(key string, value interface{}) {
	(*models.NestedStruct)(s).StructField().Self().StoreSet(key, value)
}

// StructField returns the struct field related with the nested struct.
// It differs from the Attribute method. The StructField method returns
// StructField on any level where the nested structure is defined.
func (s *NestedStruct) StructField() *StructField {
	return (*StructField)((*models.NestedStruct)(s).StructField().Self())
}

// Fields returns all nested fields within the nested struct.
func (s *NestedStruct) Fields() []*NestedField {
	var fields []*NestedField

	for _, field := range (*models.NestedStruct)(s).Fields() {
		fields = append(fields, (*NestedField)(field))
	}

	return fields
}

// Type returns the reflect.Type of the model within the nested struct.
func (s *NestedStruct) Type() reflect.Type {
	return (*models.NestedStruct)(s).Type()
}

// NestedField is the nested field within the NestedStruct.
type NestedField models.NestedField

// StructField gets the structField for provided NestedFieldvalue.
func (s *NestedField) StructField() *StructField {
	return (*StructField)((*models.NestedField)(s).StructField())
}

// StoreSet sets into the store the value 'value' for given 'key'.
func (s *NestedField) StoreSet(key string, value interface{}) {
	(*models.NestedField)(s).Self().StoreSet(key, value)
}

// StoreGet gets the value from the store at the key: 'key'.
func (s *NestedField) StoreGet(key string) (interface{}, bool) {
	return (*models.NestedField)(s).Self().StoreGet(key)
}
