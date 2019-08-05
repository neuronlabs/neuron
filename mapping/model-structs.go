package mapping

import (
	"reflect"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/namer"

	"github.com/neuronlabs/neuron-core/internal/models"
)

// ModelStruct is the structure definition for the imported models.
// It contains all the collection name, fields, config, store and a model type.
type ModelStruct models.ModelStruct

// Attr returns the attribute for the provided ModelStruct.
// If the attribute doesn't exists returns nil field and false.
func (m *ModelStruct) Attr(attr string) (*StructField, bool) {
	s, ok := m.internal().Attribute(attr)
	if !ok {
		return nil, ok
	}
	return (*StructField)(s), true
}

// Collection returns model's collection.
func (m *ModelStruct) Collection() string {
	return (*models.ModelStruct)(m).Collection()
}

// Config gets the model's defined confgi.ModelConfig.
func (m *ModelStruct) Config() *config.ModelConfig {
	return (*models.ModelStruct)(m).Config()
}

// ForeignKey checks and returns model's foreign key field.
// The 'fk' foreign key field name may be a Neuron name or Golang StructField name.
func (m *ModelStruct) ForeignKey(fk string) (*StructField, bool) {
	s, ok := m.internal().ForeignKey(fk)
	if !ok {
		return nil, ok
	}
	return (*StructField)(s), ok
}

// FilterKey checks and returns model's filter key.
// The 'fk' filter key field name may be a Neuron name or Golang StructField name.
func (m *ModelStruct) FilterKey(fk string) (*StructField, bool) {
	s, ok := m.internal().FilterKey(fk)
	if !ok {
		return nil, ok
	}
	return (*StructField)(s), ok
}

// FieldByName gets the StructField by the 'name' argument.
// The 'name' may be a StructField's Name or NeuronName.
func (m *ModelStruct) FieldByName(name string) (*StructField, bool) {
	field := m.internal().FieldByName(name)
	if field == nil {
		return nil, false
	}
	return (*StructField)(field), true
}

// Fields gets all attributes and relationships StructFields for the Model.
func (m *ModelStruct) Fields() (fields []*StructField) {
	for _, field := range (*models.ModelStruct)(m).Fields() {
		fields = append(fields, (*StructField)(field))
	}
	return fields
}

// LanguageField returns model's language field.
func (m *ModelStruct) LanguageField() *StructField {
	return (*StructField)(m.internal().LanguageField())
}

// MaxIncludedDepth gets maximum number of included nested depth.
func (m *ModelStruct) MaxIncludedDepth() int {
	return m.internal().MaxIncludedCount()
}

// NamerFunc returns the namer func used by the given model.
func (m *ModelStruct) NamerFunc() namer.Namer {
	return (namer.Namer)(m.internal().NamerFunc())
}

// Primary returns model's primary field StructField.
func (m *ModelStruct) Primary() *StructField {
	return (*StructField)(m.internal().PrimaryField())
}

// RelationField gets the relationship field for the provided string
// The 'rel' relationship field name may be a Neuron or Golang StructField name.
// If the relationship field doesn't exists returns nil and false
func (m *ModelStruct) RelationField(rel string) (*StructField, bool) {
	s, ok := m.internal().RelationshipField(rel)
	if !ok {
		return nil, ok
	}
	return (*StructField)(s), true
}

// RelationFields gets all model's relationship fields.
func (m *ModelStruct) RelationFields() []*StructField {
	modelRelations := m.internal().RelationshipFields()

	output := make([]*StructField, len(modelRelations))
	for i, rel := range modelRelations {
		output[i] = (*StructField)(rel)
	}
	return output
}

// StoreSet sets into the store the value 'value' for given 'key'.
func (m *ModelStruct) StoreSet(key interface{}, value interface{}) {
	(*models.ModelStruct)(m).StoreSet(key, value)
}

// StoreGet gets the value from the store at the key: 'key'.
func (m *ModelStruct) StoreGet(key interface{}) (interface{}, bool) {
	return (*models.ModelStruct)(m).StoreGet(key)
}

// StoreDelete deletes the store's value at 'key'.
func (m *ModelStruct) StoreDelete(key interface{}) {
	(*models.ModelStruct)(m).StoreDelete(key)
}

// StructFields return all struct fields mapping used by the model.
func (m *ModelStruct) StructFields() (mFields []*StructField) {
	for _, f := range m.internal().StructFields() {
		mFields = append(mFields, (*StructField)(f))
	}
	return mFields
}

// Type returns model's reflect.Type.
func (m *ModelStruct) Type() reflect.Type {
	return (*models.ModelStruct)(m).Type()
}

func (m *ModelStruct) internal() *models.ModelStruct {
	return (*models.ModelStruct)(m)
}
