package gormrepo

import (
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi"
	"reflect"
)

var jsIndexKey = "jsonapi.Scope/Index"

func (g *GORMRepository) setJScope(jScope *jsonapi.Scope, db *gorm.DB) {
	*db = *db.Set(jsIndexKey, jScope)
}

func (g *GORMRepository) getJScope(scope *gorm.Scope) (*jsonapi.Scope, bool) {

	v, ok := scope.Get(jsIndexKey)
	if !ok {
		g.log().Debugf("Getting jsonapi.Scope for db: %p failed.", scope.DB())
		return nil, ok
	}
	var jScope *jsonapi.Scope
	if v != nil {
		jScope, ok = v.(*jsonapi.Scope)
	}

	return jScope, ok
}

func (g *GORMRepository) getSelectedGormFieldValues(
	mStruct *gorm.ModelStruct,
	fields ...*jsonapi.StructField,
) (names []string) {

outer:
	for _, jsField := range fields {
		if jsField.IsRelationship() {
			continue
		}
		for _, gField := range mStruct.StructFields {
			if isFieldEqual(gField, jsField) {
				names = append(names, gField.DBName)
				continue outer
			}
		}
	}
	return
}

func getUpdatedGormFieldNames(
	mStruct *gorm.ModelStruct,
	scope *jsonapi.Scope,
) (res []string) {
	var fields []*jsonapi.StructField
	if len(scope.SelectedFields) == 0 {
		fields = scope.Struct.GetFields()
	} else {
		fields = scope.SelectedFields
	}

outer:
	for _, jsField := range fields {
		for _, gField := range mStruct.StructFields {
			if jsField.GetFieldIndex() == gField.Struct.Index[0] {
				if jsField.IsRelationship() {
					continue
				}
				// 	if gField.Relationship == nil {
				// 		scope.Log().Errorf("JSField is a relationship, but nor is gField. Struct: %s, field: %s", mStruct.ModelType.Name(), gField.Name)
				// 		continue outer
				// 	}

				// 	switch gField.Relationship.Kind {
				// 	case annotationBelongsTo:
				// 		res = append(res, gField.Relationship.ForeignDBNames[0])
				// 	case annotationHasMany:
				// 	case annotationHasOne:
				// 	case annotationManyToMany:
				// 	}
				// } else {
				if !jsField.IsPrimary() {
					res = append(res, gField.DBName)
				}
				// }
				continue outer
			}
		}
	}
	return
}

func (g *GORMRepository) getUpdatedFieldValues(
	mStruct *gorm.ModelStruct,
	scope *jsonapi.Scope,
) (values map[string]interface{}) {
	values = map[string]interface{}{}
	v := reflect.ValueOf(scope.Value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
outer:
	for _, jsField := range scope.SelectedFields {
		if jsField.IsRelationship() {
			continue outer
		}
		if jsField.IsPrimary() {
			continue outer
		}
		for _, gField := range mStruct.StructFields {
			if jsField.GetFieldIndex() == gField.Struct.Index[0] {

				field := v.FieldByIndex(jsField.GetReflectStructField().Index)
				fieldVal := field.Interface()
				if jsField.GetFieldKind() == jsonapi.ForeignKey {
					if reflect.DeepEqual(fieldVal, reflect.Zero(field.Type()).Interface()) {
						fieldVal = nil
					}
				}
				values[gField.DBName] = fieldVal

				continue outer
			}
		}
	}

	return
}

func checkFieldTypes(one, two reflect.Type) bool {
	switch one.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int8, reflect.Int64, reflect.Int32:
		switch two.Kind() {
		case reflect.Int, reflect.Int16, reflect.Int8, reflect.Int64, reflect.Int32:
			return true
		}
		return false
	case reflect.Uint, reflect.Uint16, reflect.Uint8, reflect.Uint64, reflect.Uint32:
		switch two.Kind() {
		case reflect.Uint, reflect.Uint16, reflect.Uint8, reflect.Uint64, reflect.Uint32:
			return true
		}
	case reflect.String:
		if two.Kind() == reflect.String {
			return true
		}
		return false
	}
	return false

}
