package gormrepo

import (
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi"
)

func getUpdatedGormFields(
	mStruct *gorm.ModelStruct,
	scope *jsonapi.Scope,
) (res []*gorm.StructField) {
outer:
	for _, jsField := range scope.SelectedFields {
		for _, gField := range mStruct.StructFields {
			if jsField.GetFieldIndex() == gField.Struct.Index[0] {
				res = append(res, gField)
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
outer:
	for _, jsField := range scope.SelectedFields {
		for _, gField := range mStruct.StructFields {
			if jsField.GetFieldIndex() == gField.Struct.Index[0] {
				if jsField.IsRelationship() {
					if gField.Relationship == nil {
						scope.Log().Errorf("JSField is a relationship, but nor is gField. Struct: %s, field: %s", mStruct.ModelType.Name(), gField.Name)
						continue outer
					}

					switch gField.Relationship.Kind {
					case annotationBelongsTo:
						res = append(res, gField.Relationship.ForeignDBNames[0])
					case annotationHasMany:
					case annotationHasOne:
					case annotationManyToMany:
					}
				} else {
					res = append(res, gField.DBName)
				}
				continue outer
			}
		}
	}
	return
}
