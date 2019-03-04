package gormrepo

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi/query/scope"
	"reflect"
	debugStack "runtime/debug"
)

var (
	IErrInvalidForeignKeyField      = errors.New("Invalid foreign key value")
	IErrNoForeignKeyField           = errors.New("The foreign key field is does not exists.")
	IErrNoAssociatedForeignKeyField = errors.New("The associated foreign key field does not exists")
	IErrInvalidRelatinoshipType     = errors.New("Invalid relationship type")
	IErrInvalidRelationshipValue    = errors.New("Provided relationship with invalid value")
)

type relationshipType int

const (
	belongsTo relationshipType = iota
	hasOne
	hasMany
	many2many
)

type modelWithFields struct {
	model  interface{}
	fields []string
}

type relationshipFieldValue struct {
	relationship relationshipType
	table        string

	relPrimaryField *gorm.StructField
	relPrimaries    []interface{}

	foreignKeyDBName string
	foreignKeyValue  interface{}

	jointableHandler gorm.JoinTableHandlerInterface

	associatedField *gorm.StructField
}

func (g *GORMRepository) prepareRelationshipScopes(
	s *scope.Scope,
) (updateRelationships []*relationshipFieldValue, err error) {
	if s.Value == nil {
		err = IErrNoValuesProvided
		return
	}

	defer func() {
		if r := recover(); r != nil {
			debugStack.PrintStack()
			switch perr := r.(type) {
			case *reflect.ValueError:
				err = fmt.Errorf("Provided invalid value input to the repository. Error: %s", perr.Error())
			case error:
				err = perr
			case string:
				err = errors.New(perr)
			default:
				err = fmt.Errorf("Unknown panic occured during getting s's relationship.")
			}
		}
	}()

	gormScope := g.db.NewScope(s.Value)

	for _, field := range gormScope.Fields() {
		if rel := field.Relationship; rel != nil {
			switch rel.Kind {
			case associationBelongsTo:
				// Association belongs to - just copy the id from the field into the foreign key
				// field

				fieldValue := field.Field

				relationScope := g.db.NewScope(fieldValue.Interface())

				idField := relationScope.PrimaryField()

				idFieldValue := idField.Field

				// dereference ptr
				if idFieldValue.Kind() == reflect.Ptr {
					idFieldValue = idFieldValue.Elem()
				}

				// copy the id from the field to the fk field
				fkField, ok := gormScope.FieldByName(rel.ForeignFieldNames[0])
				if !ok {
					err = IErrNoForeignKeyField
					return
				}

				// dereference ptr type foreignkey field
				fkValue := fkField.Field
				if fkValue.Kind() == reflect.Ptr {
					fkValue = fkValue.Elem()
				}

				// Check if type match
				if idField.Field.Type() != fkValue.Type() {
					err = IErrInvalidForeignKeyField
					return
				}
				// Set the foreignkey
				fkValue.Set(idFieldValue)

				// set the field to be zero type
				field.Field.Set(reflect.Zero(field.Field.Type()))

			case associationHasOne:
				// has one relation ship has foreign key in the different model

				// check if not nil
				if field.Field.IsNil() {
					continue
				}

				associatedField, ok := gormScope.FieldByName(rel.AssociationForeignFieldNames[0])
				if !ok {
					err = IErrNoAssociatedForeignKeyField
					return
				}

				// Check if the primary field is zero value
				relScope := g.db.NewScope(field)
				if relScope.PrimaryKeyZero() {
					continue
				}

				// if id value is valid set the relationshipFieldValue
				rfv := &relationshipFieldValue{
					relationship:    hasOne,
					table:           relScope.TableName(),
					relPrimaryField: relScope.PrimaryField().StructField,
					relPrimaries:    []interface{}{relScope.PrimaryKeyValue()},

					foreignKeyDBName: rel.ForeignDBNames[0],
					associatedField:  associatedField.StructField,
				}

				updateRelationships = append(updateRelationships, rfv)

				// Set relationship field to nil
				field.Field.Set(reflect.Zero(field.Field.Type()))

			case associationHasMany:
				if field.Field.Len() == 0 {
					continue
				}

				// get association structfield
				associatedField, ok := gormScope.FieldByName(rel.AssociationForeignFieldNames[0])
				if !ok {
					err = IErrNoAssociatedForeignKeyField
					return
				}

				t := field.Field.Type().Elem()

				if t.Kind() == reflect.Ptr {
					t = t.Elem()
				}

				relScope := g.db.NewScope(reflect.New(t).Interface())
				rfv := &relationshipFieldValue{
					relationship:    hasMany,
					table:           relScope.TableName(),
					relPrimaryField: relScope.PrimaryField().StructField,

					foreignKeyDBName: rel.ForeignDBNames[0],
					associatedField:  associatedField.StructField,
				}

			RelLoop:
				for i := 0; i < field.Field.Len(); i++ {
					elem := field.Field.Index(i)

					// Check the type
					if elem.Kind() != reflect.Ptr {
						err = IErrInvalidRelatinoshipType
						return
					}

					// Check if Nil
					if elem.IsNil() {
						continue RelLoop
					}

					// Check if contains ID value
					relScope = g.db.NewScope(elem)
					if relScope.PrimaryKeyZero() {
						continue RelLoop
					}

					rfv.relPrimaries = append(rfv.relPrimaries, relScope.PrimaryKeyValue())
				}
				updateRelationships = append(updateRelationships, rfv)
				// Set zero value
				field.Field.Set(reflect.Zero(field.Field.Type()))
			case associationManyToMany:
				if field.Field.Len() == 0 {
					continue
				}

				// get association structfield
				associatedField, ok := gormScope.FieldByName(rel.AssociationForeignFieldNames[0])
				if !ok {
					err = IErrNoAssociatedForeignKeyField
					return
				}

				t := field.Field.Type().Elem()

				if t.Kind() == reflect.Ptr {
					t = t.Elem()
				}
				// relScope := g.db.NewScope(reflect.New(t).Interface())
				rfv := &relationshipFieldValue{
					relationship:     many2many,
					associatedField:  associatedField.StructField,
					jointableHandler: rel.JoinTableHandler,
				}

				// iterate over relationship fields
			Many2Many:
				for i := 0; i < field.Field.Len(); i++ {
					single := field.Field.Index(i)
					if single.Kind() != reflect.Ptr {
						err = IErrInvalidRelatinoshipType
						return
					}

					if single.IsNil() {
						continue Many2Many
					}

					relScope := g.db.NewScope(single)
					if relScope.PrimaryKeyZero() {
						continue Many2Many
					}

					rfv.relPrimaries = append(rfv.relPrimaries, relScope.PrimaryKeyValue())
				}
				// Clear relationship field
				field.Field.Set(reflect.Zero(field.Field.Type()))

				if len(rfv.relPrimaries) == 0 {
					continue
				}
				updateRelationships = append(updateRelationships, rfv)
			}
		}

	}

	return
}

func isZero(fieldValue reflect.Value) bool {
	return reflect.DeepEqual(fieldValue, reflect.Zero(fieldValue.Type()))
}
