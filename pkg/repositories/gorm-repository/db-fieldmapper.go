package gormrepo

import (
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi/pkg/mapping"
	"github.com/kucjac/jsonapi/pkg/query/scope"

	"reflect"
)

var jsIndexKey = "scope.Scope/Index"

func (g *GORMRepository) setJScope(jScope *scope.Scope, db *gorm.DB) {
	*db = *db.Set(jsIndexKey, jScope)
}

func (g *GORMRepository) getJScope(s *gorm.Scope) (*scope.Scope, bool) {

	v, ok := s.Get(jsIndexKey)
	if !ok {
		g.log().Debugf("Getting scope.Scope for db: %p failed.", s.DB())
		return nil, ok
	}
	var jScope *scope.Scope
	if v != nil {
		jScope, ok = v.(*scope.Scope)
	}

	return jScope, ok
}

func (g *GORMRepository) getSelectedGormFieldValues(
	mStruct *gorm.ModelStruct,
	fields ...*mapping.StructField,
) (names []string) {

outer:
	for _, jsField := range fields {
		if jsField.Relationship() == nil {
			continue
		}
		for _, gField := range mStruct.StructFields {
			if gField.IsIgnored {
				continue
			}
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
	s *scope.Scope,
) (res []string) {
	var fields []*mapping.StructField
	if len(s.SelectedFields()) == 0 {
		fields = s.Struct().Fields()
	} else {
		fields = s.SelectedFields()
	}

outer:
	for _, jsField := range fields {
		for _, gField := range mStruct.StructFields {
			if gField.IsIgnored {
				continue
			}
			if jsField.ReflectField().Index[0] == gField.Struct.Index[0] {
				if jsField.Relationship() == nil {
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
				if jsField.FieldKind() != mapping.KindPrimary {
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
	s *scope.Scope,
) (values map[string]interface{}) {
	values = map[string]interface{}{}
	v := reflect.ValueOf(s.Value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
outer:
	for _, jsField := range s.SelectedFields() {
		if jsField.Relationship() != nil {
			continue outer
		}
		if jsField.FieldKind() == mapping.KindPrimary {
			continue outer
		}
	gormFields:
		for _, gField := range mStruct.StructFields {
			if gField.IsIgnored {
				continue gormFields
			}

			if isFieldEqual(gField, jsField) {
				field := v.FieldByIndex(jsField.ReflectField().Index)
				fieldVal := field.Interface()
				if jsField.FieldKind() == mapping.KindForeignKey {
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
