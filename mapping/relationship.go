package mapping

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/annotation"
	"github.com/neuronlabs/neuron-core/class"
)

// RelationshipKind is the enum used to define the Relationship's kind.
type RelationshipKind int

const (
	// RelUnknown is the unknown default relationship kind. States for the relationship internal errors.
	RelUnknown RelationshipKind = iota
	// RelBelongsTo is the enum value for the 'Belongs To' relationship.
	// This relationship kind states that the model containing the relationship field
	// contains also the foreign key of the related models.
	// The foreign key is a related model's primary field.
	RelBelongsTo
	// RelHasOne is the enum value for the 'Has One' relationship.
	// This relationship kind states that the model is in a one to one relationship with
	// the related model. It also states that the foreign key is located in the related model.
	RelHasOne
	// RelHasMany is the enum value for the 'Has Many' relationship.
	// This relationship kind states that the model is in a many to one relationship with the
	// related model. It also states that the foreign key is located in the related model.
	RelHasMany
	// RelMany2Many is the enum value for the 'Many To Many' relationship.
	// This relationship kind states that the model is in a many to many relationship with the
	// related model. This relationship requires the usage of the join model structure that contains
	// foreign keys of both related model types. The 'Relationship' struct foreign key should relate to the
	// model where the related field is stored - i.e. model 'user' has relationship field 'pets' to the model 'pet'
	// then the relationship pets foreign key should be a 'user id'. In order to get the foreign key of the related model
	// the relationship has also a field 'MtmForeignKey' which should be a 'pet id'.
	RelMany2Many
)

// String implements fmt.Stringer interface.
func (r RelationshipKind) String() string {
	switch r {
	case RelUnknown:
	case RelBelongsTo:
		return "BelongsTo"
	case RelHasOne:
		return "HasOne"
	case RelHasMany:
		return "HasMany"
	case RelMany2Many:
		return "Many2Many"
	}
	return "Unknown"
}

// Strategy is the strategy for the relationship.
// It's 'QueryOrder' is the querying order for multiple strategies.
// It defines which strategies should be used firsts.
// The 'OnError' error strategy defines how the relation should
// react on the error occurrence while.
type Strategy struct {
	QueryOrder uint
	OnError    ErrorStrategy
}

func (s *Strategy) parse(values map[string]string) errors.MultiError {
	// `neuron:"on_delete=order=1;on_error=fail"`
	var multiErr errors.MultiError
	for key, value := range values {
		switch strings.ToLower(key) {
		case annotation.Order:
			err := s.parseOrder(value)
			if err != nil {
				multiErr = append(multiErr, err)
			}
		case annotation.OnError:
			err := s.parseOnError(value)
			if err != nil {
				multiErr = append(multiErr, err)
			}
		default:
			err := errors.NewDetf(class.ModelRelationshipOptions, "invalid relationship strategy key: '%s'", key)
			err.SetDetails(key)
			multiErr = append(multiErr, err)
		}
	}
	return multiErr
}

func (s *Strategy) parseOrder(value string) errors.DetailedError {
	o, err := strconv.Atoi(value)
	if err != nil {
		err := errors.NewDetf(class.ModelRelationshipOptions, "relationship strategy order value is not an integer: '%s'", value)
		return err
	}

	if o < 0 {
		err := errors.NewDet(class.ModelRelationshipOptions, "relationship strategy order is lower than 0")
		return err
	}
	s.QueryOrder = uint(o)
	return nil
}

func (s *Strategy) parseOnError(value string) errors.DetailedError {
	switch strings.ToLower(value) {
	case annotation.FailOnError:
		s.OnError = Fail
	case annotation.ContinueOnError:
		s.OnError = Continue
	default:
		err := errors.NewDetf(class.ModelRelationshipOptions, "invalid on_error option: '%s'", value)
		return err
	}
	return nil
}

// DeleteStrategy is the strategy for the deletion options.
// The 'OnChange' varible defines how the relation should react
// on deletion of the root model.
type DeleteStrategy struct {
	Strategy
	OnChange ChangeStrategy
}

func (d *DeleteStrategy) parse(values map[string]string) errors.MultiError {
	var multiErr errors.MultiError
	for key, value := range values {
		switch strings.ToLower(key) {
		case annotation.Order:
			err := d.parseOrder(value)
			if err != nil {
				multiErr = append(multiErr, err)
			}
		case annotation.OnError:
			err := d.parseOnError(value)
			if err != nil {
				multiErr = append(multiErr, err)
			}
		case annotation.OnChange:
			err := d.parseOnChange(value)
			if err != nil {
				multiErr = append(multiErr, err)
			}
		default:
			err := errors.NewDetf(class.ModelRelationshipOptions, "invalid relationship strategy key: '%s'", key)
			err.SetDetails(key)
			multiErr = append(multiErr, err)
		}
	}
	return multiErr
}

func (d *DeleteStrategy) parseOnChange(value string) errors.DetailedError {
	switch value {
	case annotation.RelationSetNull:
		d.OnChange = SetNull
	case annotation.RelationCascade:
		d.OnChange = Cascade
	case annotation.RelationRestrict:
		d.OnChange = Restrict
	case annotation.RelationNoAction:
		d.OnChange = NoAction
	default:
		return errors.NewDetf(class.ModelRelationshipOptions, "relationship invalid 'on delete' option: '%s'", value)
	}
	return nil
}

// ChangeStrategy defines the option on how to process the relationship.
type ChangeStrategy int

// Relationship options
const (
	SetNull ChangeStrategy = iota
	NoAction
	Cascade
	Restrict
)

// ErrorStrategy is the query strategy for given relationship field.
type ErrorStrategy int

// Relationship error strategies constant values, with the default set to 'fail'.
const (
	Fail ErrorStrategy = iota
	Continue
)

// Relationship is the structure that contains the relation's required field's
// kind, join model (if exists) and the process option (onDelete, onUpdate) as well
// as the definition for the related model's type 'mStruct'.
type Relationship struct {
	// kind is a relationship kind
	kind RelationshipKind

	// foreignKey represtents the foreign key field
	foreignKey *StructField
	// mtmRelatedForeignKey is the foreign key of the many2many related model.
	// The field is only used on the many2many relationships
	mtmRelatedForeignKey *StructField
	// joinModel is the join model used for the many2many relationships
	joinModel *ModelStruct
	// joinModelName is the collection name of the join model.
	joinModelName string

	// onCreate is a relationship strategy for the creation process.
	onCreate Strategy
	// onPatch is a relationship strategy for the patch process.
	onPatch Strategy
	// onDelete is a relationship strategy for the deletion process.
	onDelete DeleteStrategy

	// mStruct is the relationship related model's structure
	mStruct *ModelStruct
	// modelType is the relationship's model Type
	modelType reflect.Type
}

// ForeignKey returns relationships foreign key.
func (r *Relationship) ForeignKey() *StructField {
	return r.foreignKey
}

// IsManyToMany defines if the relationship is of many to many type.
func (r *Relationship) IsManyToMany() bool {
	return r.isMany2Many()
}

// IsToMany defines if the relationship is of to many kind.
func (r *Relationship) IsToMany() bool {
	return r.isToMany()
}

// IsToOne defines if the relationship is of to one type.
func (r *Relationship) IsToOne() bool {
	return r.isToOne()
}

// JoinModel returns the join model for the given many2many relationship.
func (r *Relationship) JoinModel() *ModelStruct {
	return r.joinModel
}

// Kind returns relationship's kind.
func (r *Relationship) Kind() RelationshipKind {
	return r.kind
}

// ManyToManyForeignKey returns the foreign key of the many2many related model's.
func (r *Relationship) ManyToManyForeignKey() *StructField {
	return r.mtmRelatedForeignKey
}

// OnCreate gets the on create strategy for the relationship.
func (r *Relationship) OnCreate() *Strategy {
	return &r.onCreate
}

// OnDelete gets the on delete strategy for the relationship.
func (r *Relationship) OnDelete() *DeleteStrategy {
	return &r.onDelete
}

// OnPatch gets the on create strategy for the relationship.
func (r *Relationship) OnPatch() *Strategy {
	return &r.onPatch
}

// Struct returns relationship model *ModelStruct.
func (r *Relationship) Struct() *ModelStruct {
	return r.mStruct
}

func (r *Relationship) isToOne() bool {
	switch r.kind {
	case RelHasOne, RelBelongsTo:
		return true
	}
	return false
}

func (r *Relationship) isToMany() bool {
	switch r.kind {
	case RelHasOne, RelBelongsTo, RelUnknown:
		return false
	}
	return true
}

func (r *Relationship) isMany2Many() bool {
	return r.kind == RelMany2Many
}

// setForeignKey sets foreign key structfield.
func (r *Relationship) setForeignKey(s *StructField) {
	r.foreignKey = s
}

// SetKind sets the relationship kind
func (r *Relationship) setKind(kind RelationshipKind) {
	r.kind = kind
}
