package mapping

// Model is the interface used for getting and setting model primary values.
type Model interface {
	NeuronCollectionName() string
	// GetPrimaryKeyStringValue gets the primary key string value.
	GetPrimaryKeyStringValue() (string, error)
	// GetPrimaryKeyValue returns the primary key field value.
	GetPrimaryKeyValue() interface{}
	// GetPrimaryKeyHashableValue returns the primary key field value.
	GetPrimaryKeyHashableValue() interface{}
	// GetPrimaryKeyZeroValue gets the primary key zero (non set) value.
	GetPrimaryKeyZeroValue() interface{}
	// GetPrimaryKeyAddress
	GetPrimaryKeyAddress() interface{}
	// IsPrimaryKeyZero checks if the primary key value is zero.
	IsPrimaryKeyZero() bool
	// SetPrimaryKeyValue sets the primary key field value to 'src'.
	SetPrimaryKeyValue(src interface{}) error
	// SetPrimaryKeyStringValue sets the primary key field value from the string 'src'.
	SetPrimaryKeyStringValue(src string) error
}

// Fielder is the interface used to get and set model field values.
type Fielder interface {
	// GetFieldZeroValue gets 'field' zero value. A zero value is an initial - non set value.
	GetFieldZeroValue(field *StructField) (interface{}, error)
	// IsFieldZero checks if the field has zero value.
	IsFieldZero(field *StructField) (bool, error)
	// SetFieldZeroValue gets 'field' zero value. A zero value is an initial - non set value.
	SetFieldZeroValue(field *StructField) error
	// GetHashableFieldValue returns hashable field value - if the function is nil - returns nil
	// If the field is []byte it would be converted to the string.
	GetHashableFieldValue(field *StructField) (interface{}, error)
	// GetFieldStringValue gets field's string value.
	GetFieldStringValue(field *StructField) (string, error)
	// GetFieldValue returns 'field' value.
	GetFieldValue(field *StructField) (interface{}, error)
	// SetFieldValue sets the 'field''s 'value'. In order to set
	SetFieldValue(field *StructField, value interface{}) error
	// GetFieldsAddress gets field's address.
	GetFieldsAddress(field *StructField) (interface{}, error)
}

// SingleRelationer is the interface used by the model with single relationship - HasOne or BelongsTo.
type SingleRelationer interface {
	// GetRelationModel gets the model for provided 'relation' field. It is used for the single relation models
	GetRelationModel(relation *StructField) (Model, error)
	// SetRelationModel sets the 'model' value in the 'relation' field.
	SetRelationModel(relation *StructField, model Model) error
}

// MultiRelationer is the interface used to operate on the model with relationship of 'many' type
// like: HasMany or Many2Many.
type MultiRelationer interface {
	// AddRelationModel adds  'model' to the given 'relation' slice.
	AddRelationModel(relation *StructField, model Model) error
	// GetRelationModels gets the model values for the 'relation' field.
	GetRelationModels(relation *StructField) ([]Model, error)
	// GetRelationModelAt gets the 'relation' single model value at 'index' in the slice.
	GetRelationModelAt(relation *StructField, index int) (Model, error)
	// GetRelationLen gets the length of the 'relation' field.
	GetRelationLen(relation *StructField) (int, error)
}

// NewModel creates new model instance.
func NewModel(m *ModelStruct) Model {
	return NewValueSingle(m).(Model)
}

// ZeroChecker is the interface that allows to check if the value is zero.
type ZeroChecker interface {
	IsZero() bool
	GetZero() interface{}
}
