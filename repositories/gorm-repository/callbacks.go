package gormrepo

import (
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi"
	"reflect"
	"strings"
)

func isFieldEqual(gField *gorm.StructField, jField *jsonapi.StructField) bool {
	switch len(gField.Struct.Index) {
	case 1:
		return gField.Struct.Index[0] == jField.GetReflectStructField().Index[0]
	default:
		if len(jField.GetReflectStructField().Index) != len(gField.Struct.Index) {
			return false
		}

		return gField.Name == jField.GetFieldName()
	}
	return false
}

func (g *GORMRepository) changeableField(
	scope *gorm.Scope,
	field *gorm.Field,
	jScope *jsonapi.Scope,
) bool {
	if jScope != nil {

		for _, jField := range jScope.Struct.GetFields() {
			if !isFieldEqual(field.StructField, jField) {
				continue
			}

			// scope.Log("Found for field: %v", field.Name)
			if rel := jField.GetRelationship(); rel != nil {
				// scope.Log(fmt.Sprintf("Field: %v is relationship", field.Name))
				switch rel.Kind {
				case jsonapi.RelBelongsTo:
					// if relation is of belongs to kind do nothing
					break
				case jsonapi.RelHasMany, jsonapi.RelHasOne:
					if rel.Sync != nil && !*rel.Sync {
						// if the relation is not synced allow it to get locally relations
						break
					}
					return false
				case jsonapi.RelMany2Many:
					if rel.Sync != nil && *rel.Sync {
						// if the relation is synced the relationship values wouldbe taken from the
						// relationship repository
						return false
					}
					// otherwise allow to
					break
				}
			}
		}

	}

	if selectAttrs := scope.SelectAttrs(); len(selectAttrs) > 0 {
		for _, attr := range selectAttrs {
			if field.Name == attr || field.DBName == attr {
				return true
			}
		}
		return false
	}

	for _, attr := range scope.OmitAttrs() {
		if field.Name == attr || field.DBName == attr {
			return false
		}
	}

	return true
}

func (g *GORMRepository) saveAssociationCheck(
	scope *gorm.Scope,
	field *gorm.Field,
	jScope *jsonapi.Scope,
) (
	autoUpdate bool, autoCreate bool, saveReference bool,
	r *gorm.Relationship,
) {
	checkTruth := func(value interface{}) bool {
		if v, ok := value.(bool); ok && !v {
			return false
		}

		if v, ok := value.(string); ok {
			v = strings.ToLower(v)
			if v == "false" || v != "skip" {
				return false
			}
		}

		return true
	}
	changeable := g.changeableField(scope, field, jScope)
	// scope.Log(fmt.Sprintf("Field %s, changeable: %v. JSONAPI: %v", field.Name, changeable, jScope != nil))
	if changeable && !field.IsBlank && !field.IsIgnored {
		if r = field.Relationship; r != nil {

			autoUpdate, autoCreate, saveReference = true, true, true

			if value, ok := scope.Get("gorm:save_associations"); ok {
				autoUpdate = checkTruth(value)
				autoCreate = autoUpdate
			} else if value, ok := field.TagSettings["SAVE_ASSOCIATIONS"]; ok {
				autoUpdate = checkTruth(value)
				autoCreate = autoUpdate
			}

			if value, ok := scope.Get("gorm:association_autoupdate"); ok {
				autoUpdate = checkTruth(value)
			} else if value, ok := field.TagSettings["ASSOCIATION_AUTOUPDATE"]; ok {
				autoUpdate = checkTruth(value)
			}

			if value, ok := scope.Get("gorm:association_autocreate"); ok {
				autoCreate = checkTruth(value)
			} else if value, ok := field.TagSettings["ASSOCIATION_AUTOCREATE"]; ok {
				autoCreate = checkTruth(value)
			}

			if value, ok := scope.Get("gorm:association_save_reference"); ok {
				saveReference = checkTruth(value)
			} else if value, ok := field.TagSettings["ASSOCIATION_SAVE_REFERENCE"]; ok {
				saveReference = checkTruth(value)
			}
		}
	}

	return
}

func (g *GORMRepository) saveAfterAssociationsCallback(scope *gorm.Scope) {

	jScope, ok := g.getJScope(scope)
	if !ok {
		g.log().Debugf("jsonapi.Scope not found for the scope: %#v", scope)
	}

	if jScope != nil {
		gormType := scope.GetModelStruct().ModelType

		if jScope.Struct.GetType() != gormType {
			g.log().Warningf("Scope type doesn't match. JScope: %v, GormScope: %#v", jScope.Struct.GetType(), scope)
			jScope = nil
		}

	}

	for _, field := range scope.Fields() {
		autoUpdate, autoCreate, saveReference, relationship := g.saveAssociationCheck(scope, field, jScope)

		if relationship != nil && (relationship.Kind == "has_one" ||
			relationship.Kind == "has_many" ||
			relationship.Kind == "many_to_many") {
			value := field.Field

			switch value.Kind() {
			case reflect.Slice:
				for i := 0; i < value.Len(); i++ {
					newDB := scope.NewDB()
					elem := value.Index(i).Addr().Interface()
					newScope := newDB.NewScope(elem)

					if saveReference {
						if relationship.JoinTableHandler == nil && len(relationship.ForeignFieldNames) != 0 {
							for idx, fieldName := range relationship.ForeignFieldNames {
								associationForeignName := relationship.AssociationForeignDBNames[idx]
								if f, ok := scope.FieldByName(associationForeignName); ok {
									scope.Err(newScope.SetColumn(fieldName, f.Field.Interface()))
								}
							}
						}

						if relationship.PolymorphicType != "" {
							scope.Err(newScope.SetColumn(relationship.PolymorphicType, relationship.PolymorphicValue))
						}
					}

					if newScope.PrimaryKeyZero() {
						if autoCreate {
							scope.Err(newDB.Save(elem).Error)
						}
					} else if autoUpdate {
						scope.Err(newDB.Save(elem).Error)
					}

					if !scope.New(newScope.Value).PrimaryKeyZero() && saveReference {
						if joinTableHandler := relationship.JoinTableHandler; joinTableHandler != nil {
							scope.Err(joinTableHandler.Add(joinTableHandler, newDB, scope.Value, newScope.Value))
						}
					}
				}
			default:
				elem := value.Addr().Interface()
				newScope := scope.New(elem)

				if saveReference {
					if len(relationship.ForeignFieldNames) != 0 {
						for idx, fieldName := range relationship.ForeignFieldNames {
							associationForeignName := relationship.AssociationForeignDBNames[idx]
							if f, ok := scope.FieldByName(associationForeignName); ok {
								scope.Err(newScope.SetColumn(fieldName, f.Field.Interface()))
							}
						}
					}

					if relationship.PolymorphicType != "" {
						scope.Err(newScope.SetColumn(relationship.PolymorphicType, relationship.PolymorphicValue))
					}
				}

				if newScope.PrimaryKeyZero() {
					if autoCreate {
						g.log().Debugf("AutCreate single value for: %#v", elem)
						scope.Err(scope.NewDB().Save(elem).Error)
					}
				} else if autoUpdate {
					g.log().Debugf("AutoUpdate model: %#v", elem)
					scope.Err(scope.NewDB().Model(elem).Updates(elem).Error)
				}
			}

		}
	}
}

// // saveBeforeAssociationsCallback
// func saveBeforeAssociationsCallback(scope *gorm.Scope) {
// 	scope.Log("BeforeAssoctiationsCallback")
// 	for _, field := range scope.Fields() {
// 		relationship := field.Relationship

// 		if relationship != nil && relationship.Kind == "belongs_to" {
// 			scope.Log(fmt.Sprintf("Checking field: %s", field.Name))
// 			fieldValue := field.Field.Addr().Interface()
// 			newScope := scope.New(fieldValue)

// 			if newScope.PrimaryKeyZero() {
// 				scope.Log("PK is zero")
// 				continue
// 				// scope.Log("Provided invalid relationship value")
// 				// scope.Err(IErrInvalidRelationshipValue)
// 			}

// 			if len(relationship.ForeignFieldNames) != 0 {
// 				// set value's foreign key
// 				for idx, fieldName := range relationship.ForeignFieldNames {
// 					associationForeignName := relationship.AssociationForeignDBNames[idx]
// 					if foreignField, ok := scope.New(fieldValue).FieldByName(associationForeignName); ok {
// 						scope.Log(fmt.Sprintf("Setting value: %v for field: %v", foreignField.Field.Interface(), fieldName))
// 						scope.Err(scope.SetColumn(fieldName, foreignField.Field.Interface()))
// 					}
// 				}
// 			}

// 		}
// 	}
// }
