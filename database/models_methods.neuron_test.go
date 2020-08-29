// Code generated by neuron/generator. DO NOT EDIT.
// This file was generated at:
// Thu, 27 Aug 2020 15:31:09 +0200

package database

import (
	"strconv"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// Neuron_Models stores all generated models in this package.
var Neuron_Models = []mapping.Model{
	&TestModel{},
	&SoftDeletableNoHooks{},
	&Updateable{},
}

// Compile time check if SoftDeletable implements mapping.Model interface.
var _ mapping.Model = &TestModel{}

// NeuronCollectionName implements mapping.Model interface method.
// Returns the name of the collection for the 'SoftDeletable'.
func (s *TestModel) NeuronCollectionName() string {
	return "soft_deletables"
}

// IsPrimaryKeyZero implements mapping.Model interface method.
func (s *TestModel) IsPrimaryKeyZero() bool {
	return s.ID == 0
}

// GetPrimaryKeyValue implements mapping.Model interface method.
func (s *TestModel) GetPrimaryKeyValue() interface{} {
	return s.ID
}

// GetPrimaryKeyStringValue implements mapping.Model interface method.
func (s *TestModel) GetPrimaryKeyStringValue() (string, error) {
	return strconv.FormatInt(int64(s.ID), 10), nil
}

// GetPrimaryKeyAddress implements mapping.Model interface method.
func (s *TestModel) GetPrimaryKeyAddress() interface{} {
	return &s.ID
}

// GetPrimaryKeyHashableValue implements mapping.Model interface method.
func (s *TestModel) GetPrimaryKeyHashableValue() interface{} {
	return s.ID
}

// GetPrimaryKeyZeroValue implements mapping.Model interface method.
func (s *TestModel) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

// SetPrimaryKey implements mapping.Model interface method.
func (s *TestModel) SetPrimaryKeyValue(value interface{}) error {
	if v, ok := value.(int); ok {
		s.ID = v
		return nil
	}
	// Check alternate types for given field.
	switch valueType := value.(type) {
	case int8:
		s.ID = int(valueType)
	case int16:
		s.ID = int(valueType)
	case int32:
		s.ID = int(valueType)
	case int64:
		s.ID = int(valueType)
	case uint:
		s.ID = int(valueType)
	case uint8:
		s.ID = int(valueType)
	case uint16:
		s.ID = int(valueType)
	case uint32:
		s.ID = int(valueType)
	case uint64:
		s.ID = int(valueType)
	case float32:
		s.ID = int(valueType)
	case float64:
		s.ID = int(valueType)
	default:
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid value: '%T' for the primary field for model: 'SoftDeletable'", value)
	}
	return nil
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (s *TestModel) SetPrimaryKeyStringValue(value string) error {
	tmp, err := strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	if err != nil {
		return err
	}
	s.ID = int(tmp)
	return nil
}

// SetFrom implements FromSetter interface.
func (s *TestModel) SetFrom(model mapping.Model) error {
	if model == nil {
		return errors.Wrap(query.ErrInvalidInput, "provided nil model to set from")
	}
	from, ok := model.(*TestModel)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotMatch, "provided model doesn't match the input: %T", model)
	}
	*s = *from
	return nil
}

// Compile time check if SoftDeletable implements mapping.Fielder interface.
var _ mapping.Fielder = &TestModel{}

// GetFieldsAddress gets the address of provided 'field'.
func (s *TestModel) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &s.ID, nil
	case 1: // CreatedAt
		return &s.CreatedAt, nil
	case 2: // UpdatedAt
		return &s.UpdatedAt, nil
	case 3: // DeletedAt
		return &s.DeletedAt, nil
	case 4: // FieldSetBefore
		return &s.FieldSetBefore, nil
	case 5: // FieldSetAfter
		return &s.FieldSetAfter, nil
	case 6: // Integer
		return &s.Integer, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: SoftDeletable'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (s *TestModel) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return 0, nil
	case 1: // CreatedAt
		return time.Time{}, nil
	case 2: // UpdatedAt
		return time.Time{}, nil
	case 3: // DeletedAt
		return nil, nil
	case 4: // FieldSetBefore
		return "", nil
	case 5: // FieldSetAfter
		return "", nil
	case 6: // Integer
		return 0, nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (s *TestModel) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return s.ID == 0, nil
	case 1: // CreatedAt
		return s.CreatedAt == time.Time{}, nil
	case 2: // UpdatedAt
		return s.UpdatedAt == time.Time{}, nil
	case 3: // DeletedAt
		return s.DeletedAt == nil, nil
	case 4: // FieldSetBefore
		return s.FieldSetBefore == "", nil
	case 5: // FieldSetAfter
		return s.FieldSetAfter == "", nil
	case 6: // Integer
		return s.Integer == 0, nil
	}
	return false, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (s *TestModel) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		s.ID = 0
	case 1: // CreatedAt
		s.CreatedAt = time.Time{}
	case 2: // UpdatedAt
		s.UpdatedAt = time.Time{}
	case 3: // DeletedAt
		s.DeletedAt = nil
	case 4: // FieldSetBefore
		s.FieldSetBefore = ""
	case 5: // FieldSetAfter
		s.FieldSetAfter = ""
	case 6: // Integer
		s.Integer = 0
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (s *TestModel) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return s.ID, nil
	case 1: // CreatedAt
		return s.CreatedAt, nil
	case 2: // UpdatedAt
		return s.UpdatedAt, nil
	case 3: // DeletedAt
		if s.DeletedAt == nil {
			return nil, nil
		}
		return *s.DeletedAt, nil
	case 4: // FieldSetBefore
		return s.FieldSetBefore, nil
	case 5: // FieldSetAfter
		return s.FieldSetAfter, nil
	case 6: // Integer
		return s.Integer, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: 'SoftDeletable'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (s *TestModel) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return s.ID, nil
	case 1: // CreatedAt
		return s.CreatedAt, nil
	case 2: // UpdatedAt
		return s.UpdatedAt, nil
	case 3: // DeletedAt
		return s.DeletedAt, nil
	case 4: // FieldSetBefore
		return s.FieldSetBefore, nil
	case 5: // FieldSetAfter
		return s.FieldSetAfter, nil
	case 6: // Integer
		return s.Integer, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: SoftDeletable'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (s *TestModel) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if v, ok := value.(int); ok {
			s.ID = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			s.ID = int(v)
		case int16:
			s.ID = int(v)
		case int32:
			s.ID = int(v)
		case int64:
			s.ID = int(v)
		case uint:
			s.ID = int(v)
		case uint8:
			s.ID = int(v)
		case uint16:
			s.ID = int(v)
		case uint32:
			s.ID = int(v)
		case uint64:
			s.ID = int(v)
		case float32:
			s.ID = int(v)
		case float64:
			s.ID = int(v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 1: // CreatedAt
		if v, ok := value.(time.Time); ok {
			s.CreatedAt = v
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 2: // UpdatedAt
		if v, ok := value.(time.Time); ok {
			s.UpdatedAt = v
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 3: // DeletedAt
		if value == nil {
			s.DeletedAt = nil
			return nil
		}
		if v, ok := value.(*time.Time); ok {
			s.DeletedAt = v
			return nil
		}
		// Check if it is non-pointer value.
		if v, ok := value.(time.Time); ok {
			s.DeletedAt = &v
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 4: // FieldSetBefore
		if v, ok := value.(string); ok {
			s.FieldSetBefore = v
			return nil
		}

		// Check alternate types for the FieldSetBefore.
		if v, ok := value.([]byte); ok {
			s.FieldSetBefore = string(v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 5: // FieldSetAfter
		if v, ok := value.(string); ok {
			s.FieldSetAfter = v
			return nil
		}

		// Check alternate types for the FieldSetAfter.
		if v, ok := value.([]byte); ok {
			s.FieldSetAfter = string(v)
			return nil
		}
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 6: // Integer
		if v, ok := value.(int); ok {
			s.Integer = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			s.Integer = int(v)
		case int16:
			s.Integer = int(v)
		case int32:
			s.Integer = int(v)
		case int64:
			s.Integer = int(v)
		case uint:
			s.Integer = int(v)
		case uint8:
			s.Integer = int(v)
		case uint16:
			s.Integer = int(v)
		case uint32:
			s.Integer = int(v)
		case uint64:
			s.Integer = int(v)
		case float32:
			s.Integer = int(v)
		case float64:
			s.Integer = int(v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for the model: 'SoftDeletable'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (s *TestModel) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 1: // CreatedAt
		temp := s.CreatedAt
		if err := s.CreatedAt.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", s.CreatedAt, err)
		}
		bt, err := s.CreatedAt.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'CreatedAt' value: '%v' to parse string. Err: %v", s.CreatedAt, err)
		}
		s.CreatedAt = temp
		return string(bt), nil
	case 2: // UpdatedAt
		temp := s.UpdatedAt
		if err := s.UpdatedAt.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'UpdatedAt' value: '%v' to parse string. Err: %v", s.UpdatedAt, err)
		}
		bt, err := s.UpdatedAt.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'UpdatedAt' value: '%v' to parse string. Err: %v", s.UpdatedAt, err)
		}
		s.UpdatedAt = temp
		return string(bt), nil
	case 3: // DeletedAt
		var base time.Time
		temp := &base
		if err := temp.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", s.DeletedAt, err)
		}
		bt, err := temp.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", s.DeletedAt, err)
		}

		return string(bt), nil
	case 4: // FieldSetBefore
		return value, nil
	case 5: // FieldSetAfter
		return value, nil
	case 6: // Integer
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: SoftDeletable'", field.Name())
}

// Compile time check if SoftDeletableNoHooks implements mapping.Model interface.
var _ mapping.Model = &SoftDeletableNoHooks{}

// NeuronCollectionName implements mapping.Model interface method.
// Returns the name of the collection for the 'SoftDeletableNoHooks'.
func (s *SoftDeletableNoHooks) NeuronCollectionName() string {
	return "soft_deletable_no_hooks"
}

// IsPrimaryKeyZero implements mapping.Model interface method.
func (s *SoftDeletableNoHooks) IsPrimaryKeyZero() bool {
	return s.ID == 0
}

// GetPrimaryKeyValue implements mapping.Model interface method.
func (s *SoftDeletableNoHooks) GetPrimaryKeyValue() interface{} {
	return s.ID
}

// GetPrimaryKeyStringValue implements mapping.Model interface method.
func (s *SoftDeletableNoHooks) GetPrimaryKeyStringValue() (string, error) {
	return strconv.FormatInt(int64(s.ID), 10), nil
}

// GetPrimaryKeyAddress implements mapping.Model interface method.
func (s *SoftDeletableNoHooks) GetPrimaryKeyAddress() interface{} {
	return &s.ID
}

// GetPrimaryKeyHashableValue implements mapping.Model interface method.
func (s *SoftDeletableNoHooks) GetPrimaryKeyHashableValue() interface{} {
	return s.ID
}

// GetPrimaryKeyZeroValue implements mapping.Model interface method.
func (s *SoftDeletableNoHooks) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

// SetPrimaryKey implements mapping.Model interface method.
func (s *SoftDeletableNoHooks) SetPrimaryKeyValue(value interface{}) error {
	if v, ok := value.(int); ok {
		s.ID = v
		return nil
	}
	// Check alternate types for given field.
	switch valueType := value.(type) {
	case int8:
		s.ID = int(valueType)
	case int16:
		s.ID = int(valueType)
	case int32:
		s.ID = int(valueType)
	case int64:
		s.ID = int(valueType)
	case uint:
		s.ID = int(valueType)
	case uint8:
		s.ID = int(valueType)
	case uint16:
		s.ID = int(valueType)
	case uint32:
		s.ID = int(valueType)
	case uint64:
		s.ID = int(valueType)
	case float32:
		s.ID = int(valueType)
	case float64:
		s.ID = int(valueType)
	default:
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid value: '%T' for the primary field for model: 'SoftDeletableNoHooks'", value)
	}
	return nil
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (s *SoftDeletableNoHooks) SetPrimaryKeyStringValue(value string) error {
	tmp, err := strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	if err != nil {
		return err
	}
	s.ID = int(tmp)
	return nil
}

// SetFrom implements FromSetter interface.
func (s *SoftDeletableNoHooks) SetFrom(model mapping.Model) error {
	if model == nil {
		return errors.Wrap(query.ErrInvalidInput, "provided nil model to set from")
	}
	from, ok := model.(*SoftDeletableNoHooks)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotMatch, "provided model doesn't match the input: %T", model)
	}
	*s = *from
	return nil
}

// Compile time check if SoftDeletableNoHooks implements mapping.Fielder interface.
var _ mapping.Fielder = &SoftDeletableNoHooks{}

// GetFieldsAddress gets the address of provided 'field'.
func (s *SoftDeletableNoHooks) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &s.ID, nil
	case 1: // DeletedAt
		return &s.DeletedAt, nil
	case 2: // Integer
		return &s.Integer, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: SoftDeletableNoHooks'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (s *SoftDeletableNoHooks) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return 0, nil
	case 1: // DeletedAt
		return nil, nil
	case 2: // Integer
		return 0, nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (s *SoftDeletableNoHooks) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return s.ID == 0, nil
	case 1: // DeletedAt
		return s.DeletedAt == nil, nil
	case 2: // Integer
		return s.Integer == 0, nil
	}
	return false, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (s *SoftDeletableNoHooks) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		s.ID = 0
	case 1: // DeletedAt
		s.DeletedAt = nil
	case 2: // Integer
		s.Integer = 0
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (s *SoftDeletableNoHooks) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return s.ID, nil
	case 1: // DeletedAt
		if s.DeletedAt == nil {
			return nil, nil
		}
		return *s.DeletedAt, nil
	case 2: // Integer
		return s.Integer, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: 'SoftDeletableNoHooks'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (s *SoftDeletableNoHooks) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return s.ID, nil
	case 1: // DeletedAt
		return s.DeletedAt, nil
	case 2: // Integer
		return s.Integer, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: SoftDeletableNoHooks'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (s *SoftDeletableNoHooks) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if v, ok := value.(int); ok {
			s.ID = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			s.ID = int(v)
		case int16:
			s.ID = int(v)
		case int32:
			s.ID = int(v)
		case int64:
			s.ID = int(v)
		case uint:
			s.ID = int(v)
		case uint8:
			s.ID = int(v)
		case uint16:
			s.ID = int(v)
		case uint32:
			s.ID = int(v)
		case uint64:
			s.ID = int(v)
		case float32:
			s.ID = int(v)
		case float64:
			s.ID = int(v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 1: // DeletedAt
		if value == nil {
			s.DeletedAt = nil
			return nil
		}
		if v, ok := value.(*time.Time); ok {
			s.DeletedAt = v
			return nil
		}
		// Check if it is non-pointer value.
		if v, ok := value.(time.Time); ok {
			s.DeletedAt = &v
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 2: // Integer
		if v, ok := value.(int); ok {
			s.Integer = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			s.Integer = int(v)
		case int16:
			s.Integer = int(v)
		case int32:
			s.Integer = int(v)
		case int64:
			s.Integer = int(v)
		case uint:
			s.Integer = int(v)
		case uint8:
			s.Integer = int(v)
		case uint16:
			s.Integer = int(v)
		case uint32:
			s.Integer = int(v)
		case uint64:
			s.Integer = int(v)
		case float32:
			s.Integer = int(v)
		case float64:
			s.Integer = int(v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for the model: 'SoftDeletableNoHooks'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (s *SoftDeletableNoHooks) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 1: // DeletedAt
		var base time.Time
		temp := &base
		if err := temp.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", s.DeletedAt, err)
		}
		bt, err := temp.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'DeletedAt' value: '%v' to parse string. Err: %v", s.DeletedAt, err)
		}

		return string(bt), nil
	case 2: // Integer
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: SoftDeletableNoHooks'", field.Name())
}

// Compile time check if Updateable implements mapping.Model interface.
var _ mapping.Model = &Updateable{}

// NeuronCollectionName implements mapping.Model interface method.
// Returns the name of the collection for the 'Updateable'.
func (u *Updateable) NeuronCollectionName() string {
	return "updateables"
}

// IsPrimaryKeyZero implements mapping.Model interface method.
func (u *Updateable) IsPrimaryKeyZero() bool {
	return u.ID == 0
}

// GetPrimaryKeyValue implements mapping.Model interface method.
func (u *Updateable) GetPrimaryKeyValue() interface{} {
	return u.ID
}

// GetPrimaryKeyStringValue implements mapping.Model interface method.
func (u *Updateable) GetPrimaryKeyStringValue() (string, error) {
	return strconv.FormatInt(int64(u.ID), 10), nil
}

// GetPrimaryKeyAddress implements mapping.Model interface method.
func (u *Updateable) GetPrimaryKeyAddress() interface{} {
	return &u.ID
}

// GetPrimaryKeyHashableValue implements mapping.Model interface method.
func (u *Updateable) GetPrimaryKeyHashableValue() interface{} {
	return u.ID
}

// GetPrimaryKeyZeroValue implements mapping.Model interface method.
func (u *Updateable) GetPrimaryKeyZeroValue() interface{} {
	return 0
}

// SetPrimaryKey implements mapping.Model interface method.
func (u *Updateable) SetPrimaryKeyValue(value interface{}) error {
	if v, ok := value.(int); ok {
		u.ID = v
		return nil
	}
	// Check alternate types for given field.
	switch valueType := value.(type) {
	case int8:
		u.ID = int(valueType)
	case int16:
		u.ID = int(valueType)
	case int32:
		u.ID = int(valueType)
	case int64:
		u.ID = int(valueType)
	case uint:
		u.ID = int(valueType)
	case uint8:
		u.ID = int(valueType)
	case uint16:
		u.ID = int(valueType)
	case uint32:
		u.ID = int(valueType)
	case uint64:
		u.ID = int(valueType)
	case float32:
		u.ID = int(valueType)
	case float64:
		u.ID = int(valueType)
	default:
		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid value: '%T' for the primary field for model: 'Updateable'", value)
	}
	return nil
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (u *Updateable) SetPrimaryKeyStringValue(value string) error {
	tmp, err := strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	if err != nil {
		return err
	}
	u.ID = int(tmp)
	return nil
}

// SetFrom implements FromSetter interface.
func (u *Updateable) SetFrom(model mapping.Model) error {
	if model == nil {
		return errors.Wrap(query.ErrInvalidInput, "provided nil model to set from")
	}
	from, ok := model.(*Updateable)
	if !ok {
		return errors.WrapDetf(mapping.ErrModelNotMatch, "provided model doesn't match the input: %T", model)
	}
	*u = *from
	return nil
}

// Compile time check if Updateable implements mapping.Fielder interface.
var _ mapping.Fielder = &Updateable{}

// GetFieldsAddress gets the address of provided 'field'.
func (u *Updateable) GetFieldsAddress(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return &u.ID, nil
	case 1: // UpdatedAt
		return &u.UpdatedAt, nil
	case 2: // Integer
		return &u.Integer, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Updateable'", field.Name())
}

// GetFieldZeroValue implements mapping.Fielder interface.s
func (u *Updateable) GetFieldZeroValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return 0, nil
	case 1: // UpdatedAt
		return time.Time{}, nil
	case 2: // Integer
		return 0, nil
	default:
		return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
}

// IsFieldZero implements mapping.Fielder interface.
func (u *Updateable) IsFieldZero(field *mapping.StructField) (bool, error) {
	switch field.Index[0] {
	case 0: // ID
		return u.ID == 0, nil
	case 1: // UpdatedAt
		return u.UpdatedAt == time.Time{}, nil
	case 2: // Integer
		return u.Integer == 0, nil
	}
	return false, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
}

// SetFieldZeroValue implements mapping.Fielder interface.s
func (u *Updateable) SetFieldZeroValue(field *mapping.StructField) error {
	switch field.Index[0] {
	case 0: // ID
		u.ID = 0
	case 1: // UpdatedAt
		u.UpdatedAt = time.Time{}
	case 2: // Integer
		u.Integer = 0
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field name: '%s'", field.Name())
	}
	return nil
}

// GetHashableFieldValue implements mapping.Fielder interface.
func (u *Updateable) GetHashableFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return u.ID, nil
	case 1: // UpdatedAt
		return u.UpdatedAt, nil
	case 2: // Integer
		return u.Integer, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: 'Updateable'", field.Name())
}

// GetFieldValue implements mapping.Fielder interface.
func (u *Updateable) GetFieldValue(field *mapping.StructField) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return u.ID, nil
	case 1: // UpdatedAt
		return u.UpdatedAt, nil
	case 2: // Integer
		return u.Integer, nil
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Updateable'", field.Name())
}

// SetFieldValue implements mapping.Fielder interface.
func (u *Updateable) SetFieldValue(field *mapping.StructField, value interface{}) (err error) {
	switch field.Index[0] {
	case 0: // ID
		if v, ok := value.(int); ok {
			u.ID = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			u.ID = int(v)
		case int16:
			u.ID = int(v)
		case int32:
			u.ID = int(v)
		case int64:
			u.ID = int(v)
		case uint:
			u.ID = int(v)
		case uint8:
			u.ID = int(v)
		case uint16:
			u.ID = int(v)
		case uint32:
			u.ID = int(v)
		case uint64:
			u.ID = int(v)
		case float32:
			u.ID = int(v)
		case float64:
			u.ID = int(v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	case 1: // UpdatedAt
		if v, ok := value.(time.Time); ok {
			u.UpdatedAt = v
			return nil
		}

		return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
	case 2: // Integer
		if v, ok := value.(int); ok {
			u.Integer = v
			return nil
		}

		switch v := value.(type) {
		case int8:
			u.Integer = int(v)
		case int16:
			u.Integer = int(v)
		case int32:
			u.Integer = int(v)
		case int64:
			u.Integer = int(v)
		case uint:
			u.Integer = int(v)
		case uint8:
			u.Integer = int(v)
		case uint16:
			u.Integer = int(v)
		case uint32:
			u.Integer = int(v)
		case uint64:
			u.Integer = int(v)
		case float32:
			u.Integer = int(v)
		case float64:
			u.Integer = int(v)
		default:
			return errors.Wrapf(mapping.ErrFieldValue, "provided invalid field type: '%T' for the field: %s", value, field.Name())
		}
		return nil
	default:
		return errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for the model: 'Updateable'", field.Name())
	}
}

// SetPrimaryKeyStringValue implements mapping.Model interface method.
func (u *Updateable) ParseFieldsStringValue(field *mapping.StructField, value string) (interface{}, error) {
	switch field.Index[0] {
	case 0: // ID
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	case 1: // UpdatedAt
		temp := u.UpdatedAt
		if err := u.UpdatedAt.UnmarshalText([]byte(value)); err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'UpdatedAt' value: '%v' to parse string. Err: %v", u.UpdatedAt, err)
		}
		bt, err := u.UpdatedAt.MarshalText()
		if err != nil {
			return "", errors.Wrapf(mapping.ErrFieldValue, "invalid field 'UpdatedAt' value: '%v' to parse string. Err: %v", u.UpdatedAt, err)
		}
		u.UpdatedAt = temp
		return string(bt), nil
	case 2: // Integer
		return strconv.ParseInt(value, 10, mapping.IntegerBitSize)
	}
	return nil, errors.Wrapf(mapping.ErrInvalidModelField, "provided invalid field: '%s' for given model: Updateable'", field.Name())
}