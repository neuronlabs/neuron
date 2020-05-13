// +build !codeanalysis

package query

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

var (
	_ mapping.Model            = &HasOneModel{}
	_ mapping.SingleRelationer = &HasOneModel{}
)

// HasOneModel is the model that have has-one relationship.
type HasOneModel struct {
	ID     int           `neuron:"type=primary"`
	HasOne *ForeignModel `neuron:"type=relation;foreign=ForeignKey"`
}

func (m *HasOneModel) NeuronCollectionName() string {
	return "has_one_models"
}

func (m *HasOneModel) GetRelationModel(relation *mapping.StructField) (mapping.Model, error) {
	return m.HasOne, nil
}

func (m *HasOneModel) SetRelationModel(relation *mapping.StructField, model mapping.Model) error {
	m.HasOne = model.(*ForeignModel)
	return nil
}

func (m *HasOneModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

func (m *HasOneModel) IsPrimaryKeyZero() bool {
	return m.ID == 0
}

func (m *HasOneModel) GetPrimaryKeyHashableValue() interface{} {
	return m.ID
}

func (m *HasOneModel) GetPrimaryKeyValue() interface{} {
	return m.ID
}

func (m *HasOneModel) SetPrimaryKeyValue(value interface{}) error {
	m.ID = value.(int)
	return nil
}

var (
	_ mapping.Model           = &HasManyModel{}
	_ mapping.MultiRelationer = &HasManyModel{}
)

// HasManyModel is the model with the has-many relationship.
type HasManyModel struct {
	ID      int             `neuron:"type=primary"`
	HasMany []*ForeignModel `neuron:"type=relation;foreign=ForeignKey"`
}

func (m *HasManyModel) GetPrimaryKeyValue() interface{} {
	return m.ID
}

func (m *HasManyModel) NeuronCollectionName() string {
	return "has_many_models"
}

func (m *HasManyModel) AddRelationModel(relation *mapping.StructField, model mapping.Model) error {
	m.HasMany = append(m.HasMany, model.(*ForeignModel))
	return nil
}

func (m *HasManyModel) GetRelationModels(relation *mapping.StructField) ([]mapping.Model, error) {
	models := make([]mapping.Model, len(m.HasMany))
	for i, model := range m.HasMany {
		models[i] = model
	}
	return models, nil
}

func (m *HasManyModel) GetRelationModelAt(relation *mapping.StructField, index int) (mapping.Model, error) {
	if index >= len(m.HasMany) {
		// TODO: change class for this entity
		return nil, errors.New(ClassInvalidModels, "index out of range")
	}
	return m.HasMany[index], nil
}

func (m *HasManyModel) GetRelationLen(relation *mapping.StructField) (int, error) {
	return len(m.HasMany), nil
}

func (m *HasManyModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

func (m *HasManyModel) IsPrimaryKeyZero() bool {
	return m.ID == 0
}

func (m *HasManyModel) GetPrimaryKeyHashableValue() interface{} {
	return m.ID
}

func (m *HasManyModel) SetPrimaryKeyValue(value interface{}) error {
	m.ID = value.(int)
	return nil
}

var (
	_ mapping.Model   = &ForeignModel{}
	_ mapping.Fielder = &ForeignModel{}
)

// ForeignModel is the model that have foreign key.
type ForeignModel struct {
	ID         int `neuron:"type=primary"`
	ForeignKey int `neuron:"type=foreign"`
}

func (f *ForeignModel) GetPrimaryKeyValue() interface{} {
	return f.ID
}

func (f *ForeignModel) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return f.ID, nil
	case "ForeignKey":
		return f.ForeignKey, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (f *ForeignModel) NeuronCollectionName() string {
	return "foreign_models"
}

func (f *ForeignModel) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Name() {
	case "ID":
		f.ID = 0
	case "ForeignKey":
		f.ForeignKey = 0
	}
	return nil
}

func (f *ForeignModel) IsPrimaryKeyZero() bool {
	return f.ID == 0
}

func (f *ForeignModel) GetPrimaryKeyHashableValue() interface{} {
	return f.ID
}

func (f *ForeignModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

func (f *ForeignModel) SetPrimaryKeyValue(value interface{}) error {
	f.ID = value.(int)
	return nil
}

func (f *ForeignModel) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Name() {
	case "ID":
		return f.ID == 0, nil
	case "ForeignKey":
		return f.ForeignKey == 0, nil
	}
	return false, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (f *ForeignModel) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return f.ID, nil
	case "ForeignKey":
		return f.ForeignKey, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (f *ForeignModel) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return 0, nil
	case "ForeignKey":
		return 0, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (f *ForeignModel) SetFieldValue(field *mapping.StructField, value interface{}) error {
	switch field.Name() {
	case "ID":
		f.ID = value.(int)
	case "ForeignKey":
		f.ForeignKey = value.(int)
	default:
		return errors.New(mapping.ClassInvalidFieldValue, "invalid field")
	}
	return nil
}

var (
	_ mapping.Model           = &ManyToManyModel{}
	_ mapping.MultiRelationer = &ManyToManyModel{}
)

// ManyToManyModel is the model with many2many relationship.
type ManyToManyModel struct {
	ID        int             `neuron:"type=primary"`
	Many2Many []*RelatedModel `neuron:"type=relation;many2many=JoinModel;foreign=ForeignKey,MtMForeignKey"`
}

func (m *ManyToManyModel) GetPrimaryKeyValue() interface{} {
	return m.ID
}

func (m *ManyToManyModel) NeuronCollectionName() string {
	return "many_to_many_models"
}

func (m *ManyToManyModel) AddRelationModel(relation *mapping.StructField, model mapping.Model) error {
	m.Many2Many = append(m.Many2Many, model.(*RelatedModel))
	return nil
}

func (m *ManyToManyModel) GetRelationModels(relation *mapping.StructField) ([]mapping.Model, error) {
	models := make([]mapping.Model, len(m.Many2Many))
	for i := range models {
		models[i] = m.Many2Many[i]
	}
	return models, nil
}

func (m *ManyToManyModel) GetRelationModelAt(relation *mapping.StructField, index int) (mapping.Model, error) {
	if index >= len(m.Many2Many) {
		// TODO: change class for this entity
		return nil, errors.New(mapping.ClassInvalidFieldValue, "index out of range")
	}
	return m.Many2Many[index], nil
}

func (m *ManyToManyModel) GetRelationLen(relation *mapping.StructField) (int, error) {
	return len(m.Many2Many), nil
}

func (m *ManyToManyModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

func (m *ManyToManyModel) IsPrimaryKeyZero() bool {
	return m.ID == 0
}

func (m *ManyToManyModel) GetPrimaryKeyHashableValue() interface{} {
	return m.ID
}

func (m *ManyToManyModel) SetPrimaryKeyValue(value interface{}) error {
	m.ID = value.(int)
	return nil
}

var (
	_ mapping.Model   = &JoinModel{}
	_ mapping.Fielder = &JoinModel{}
)

// JoinModel is the model used as a join model for the many2many relationships.
type JoinModel struct {
	ID            int `neuron:"type=primary"`
	ForeignKey    int `neuron:"type=foreign"`
	MtMForeignKey int `neuron:"type=foreign"`
}

func (m *JoinModel) GetPrimaryKeyValue() interface{} {
	return m.ID
}

func (m *JoinModel) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return m.ID, nil
	case "ForeignKey":
		return m.ForeignKey, nil
	case "MtMForeignKey":
		return m.MtMForeignKey, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *JoinModel) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Name() {
	case "ID":
		m.ID = 0
	case "ForeignKey":
		m.ForeignKey = 0
	case "MtmForeignKey":
		m.MtMForeignKey = 0
	}
	return nil
}

func (m *JoinModel) NeuronCollectionName() string {
	return "join_models"
}

func (m *JoinModel) IsPrimaryKeyZero() bool {
	return m.ID == 0
}

func (m *JoinModel) GetPrimaryKeyHashableValue() interface{} {
	return m.ID
}

func (m *JoinModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

func (m *JoinModel) SetPrimaryKeyValue(value interface{}) error {
	m.ID = value.(int)
	return nil
}

func (m *JoinModel) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Name() {
	case "ID":
		return m.ID == 0, nil
	case "ForeignKey":
		return m.ForeignKey == 0, nil
	case "MtMForeignKey":
		return m.MtMForeignKey == 0, nil
	}
	return false, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *JoinModel) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return m.ID, nil
	case "ForeignKey":
		return m.ForeignKey, nil
	case "MtMForeignKey":
		return m.MtMForeignKey, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *JoinModel) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return 0, nil
	case "ForeignKey":
		return 0, nil
	case "MtMForeignKey":
		return 0, nil
	}
	return false, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *JoinModel) SetFieldValue(field *mapping.StructField, value interface{}) error {
	switch field.Name() {
	case "ID":
		m.ID = value.(int)
	case "ForeignKey":
		m.ForeignKey = value.(int)
	case "MtMForeignKey":
		m.MtMForeignKey = value.(int)
	default:
		return errors.New(mapping.ClassInvalidFieldValue, "invalid field")
	}
	return nil
}

var (
	_ mapping.Model   = &RelatedModel{}
	_ mapping.Fielder = &RelatedModel{}
)

// RelatedModel is the related model in the many2many relationship.
type RelatedModel struct {
	ID         int     `neuron:"type=primary"`
	FloatField float64 `neuron:"type=attr"`
}

func (m *RelatedModel) GetPrimaryKeyValue() interface{} {
	return m.ID
}

func (m *RelatedModel) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return m.ID, nil
	case "FloatField":
		return m.FloatField, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *RelatedModel) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Name() {
	case "ID":
		m.ID = 0
	case "FloatField":
		m.FloatField = 0
	}
	return nil
}

func (m *RelatedModel) NeuronCollectionName() string {
	return "related_models"
}

func (m *RelatedModel) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Name() {
	case "ID":
		return m.ID == 0, nil
	case "FloatField":
		return m.FloatField == 0, nil
	}
	return false, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *RelatedModel) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return m.ID, nil
	case "FloatField":
		return m.FloatField, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *RelatedModel) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return 0, nil
	case "FloatField":
		return 0.0, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *RelatedModel) SetFieldValue(field *mapping.StructField, value interface{}) error {
	switch field.Name() {
	case "ID":
		m.ID = value.(int)
	case "FloatField":
		m.FloatField = value.(float64)
	default:
		return errors.New(mapping.ClassInvalidFieldValue, "invalid field")
	}
	return nil
}

func (m *RelatedModel) IsPrimaryKeyZero() bool {
	return m.ID == 0
}

func (m *RelatedModel) GetPrimaryKeyHashableValue() interface{} {
	return m.ID
}

func (m *RelatedModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

func (m *RelatedModel) SetPrimaryKeyValue(value interface{}) error {
	m.ID = value.(int)
	return nil
}

var (
	_ mapping.Model            = &ForeignWithRelation{}
	_ mapping.Fielder          = &ForeignWithRelation{}
	_ mapping.SingleRelationer = &ForeignWithRelation{}
)

type ForeignWithRelation struct {
	ID         int
	Relation   *HasManyWithRelation `neuron:"foreign=ForeignKey"`
	ForeignKey int                  `neuron:"type=foreign"`
}

func (m *ForeignWithRelation) GetPrimaryKeyValue() interface{} {
	return m.ID
}

func (m *ForeignWithRelation) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return m.ID, nil
	case "ForeignKey":
		return m.ForeignKey, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *ForeignWithRelation) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Name() {
	case "ID":
		m.ID = 0
	case "ForeignKey":
		m.ForeignKey = 0
	}
	return nil
}

func (m *ForeignWithRelation) NeuronCollectionName() string {
	return "foreign_with_relations"
}

func (m *ForeignWithRelation) GetRelationModel(relation *mapping.StructField) (mapping.Model, error) {
	return m.Relation, nil
}

func (m *ForeignWithRelation) SetRelationModel(relation *mapping.StructField, model mapping.Model) error {
	m.Relation = model.(*HasManyWithRelation)
	return nil
}

func (m *ForeignWithRelation) IsPrimaryKeyZero() bool {
	return m.ID == 0
}

func (m *ForeignWithRelation) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

func (m *ForeignWithRelation) GetPrimaryKeyHashableValue() interface{} {
	return m.ID
}

func (m *ForeignWithRelation) SetPrimaryKeyValue(value interface{}) error {
	m.ID = value.(int)
	return nil
}

func (m *ForeignWithRelation) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Name() {
	case "ID":
		return m.ID == 0, nil
	case "ForeignKey":
		return m.ForeignKey == 0, nil
	}
	return false, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *ForeignWithRelation) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return m.ID, nil
	case "ForeignKey":
		return m.ForeignKey, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *ForeignWithRelation) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Name() {
	case "ID":
		return 0, nil
	case "ForeignKey":
		return 0, nil
	}
	return nil, errors.New(mapping.ClassInvalidFieldValue, "invalid field")
}

func (m *ForeignWithRelation) SetFieldValue(field *mapping.StructField, value interface{}) error {
	switch field.Name() {
	case "ID":
		m.ID = value.(int)
	case "ForeignKey":
		m.ForeignKey = value.(int)
	default:
		return errors.New(mapping.ClassInvalidFieldValue, "invalid field")
	}
	return nil
}

var (
	_ mapping.Model           = &HasManyWithRelation{}
	_ mapping.MultiRelationer = &HasManyWithRelation{}
)

type HasManyWithRelation struct {
	ID       int
	Relation []*ForeignWithRelation `neuron:"foreign=ForeignKey"`
}

func (m *HasManyWithRelation) GetPrimaryKeyValue() interface{} {
	return m.ID
}

func (m *HasManyWithRelation) NeuronCollectionName() string {
	return "has_many_with_relations"
}

func (m *HasManyWithRelation) AddRelationModel(relation *mapping.StructField, model mapping.Model) error {
	m.Relation = append(m.Relation, model.(*ForeignWithRelation))
	return nil
}

func (m *HasManyWithRelation) GetRelationModels(relation *mapping.StructField) ([]mapping.Model, error) {
	models := make([]mapping.Model, len(m.Relation))
	for i := range m.Relation {
		models[i] = m.Relation[i]
	}
	return models, nil
}

func (m *HasManyWithRelation) GetRelationModelAt(relation *mapping.StructField, index int) (mapping.Model, error) {
	if index >= len(m.Relation) {
		return nil, errors.New(mapping.ClassInvalidRelationValue, "index out of range")
	}
	return m.Relation[index], nil
}

func (m *HasManyWithRelation) GetRelationLen(relation *mapping.StructField) (int, error) {
	return len(m.Relation), nil
}

func (m *HasManyWithRelation) IsPrimaryKeyZero() bool {
	return m.ID == 0
}

func (m *HasManyWithRelation) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

func (m *HasManyWithRelation) GetPrimaryKeyHashableValue() interface{} {
	return m.ID
}

func (m *HasManyWithRelation) SetPrimaryKeyValue(value interface{}) error {
	m.ID = value.(int)
	return nil
}
