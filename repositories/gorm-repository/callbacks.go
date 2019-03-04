package gormrepo

import (
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi/mapping"
	"github.com/kucjac/jsonapi/query/scope"
	"reflect"
	"strings"
)

func isFieldEqual(gField *gorm.StructField, jField *mapping.StructField) bool {
	switch len(gField.Struct.Index) {
	case 1:
		return gField.Struct.Index[0] == jField.ReflectField().Index[0]
	default:
		if len(jField.ReflectField().Index) != len(gField.Struct.Index) {
			return false
		}

		return gField.Name == jField.Name()
	}
	return false
}

func (g *GORMRepository) changeableField(
	gScope *gorm.Scope,
	field *gorm.Field,
	jScope *scope.Scope,
) bool {
	if jScope != nil {

		for _, jField := range jScope.Struct().Fields() {

			if !isFieldEqual(field.StructField, jField) {
				continue
			}

			// gScope.Log("Found for field: %v", field.Name)

			if rel := jField.Relationship(); rel != nil {
				// gScope.Log(fmt.Sprintf("Field: %v is relationship", field.Name))
				switch rel.Kind() {
				case mapping.RelBelongsTo:
					// if relation is of belongs to kind do nothing
					return false
				case mapping.RelHasMany, mapping.RelHasOne:

					if rel.Sync() != nil && !*rel.Sync() {
						// if the relation is not synced allow it to get locally relations
						break
					}
					return false
				case mapping.RelMany2Many:
					if rel.Sync() != nil && *rel.Sync() {
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

	if selectAttrs := gScope.SelectAttrs(); len(selectAttrs) > 0 {
		for _, attr := range selectAttrs {
			if field.Name == attr || field.DBName == attr {
				return true
			}
		}
		return false
	}

	for _, attr := range gScope.OmitAttrs() {
		if field.Name == attr || field.DBName == attr {
			return false
		}
	}

	return true
}

func (g *GORMRepository) saveAssociationCheck(
	gScope *gorm.Scope,
	field *gorm.Field,
	jScope *scope.Scope,
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
	changeable := g.changeableField(gScope, field, jScope)
	// gScope.Log(fmt.Sprintf("Field %s, changeable: %v. JSONAPI: %v", field.Name, changeable, jScope != nil))
	// g.log().Debugf("Field: %s changeable: %v", field.Name, changeable)
	if changeable && !field.IsBlank && !field.IsIgnored {
		if r = field.Relationship; r != nil {

			autoUpdate, autoCreate, saveReference = true, true, true
			// g.log().Debugf("AutoUpdate: %v, AutoCreate: %v, saveReference: %v", autoUpdate, autoCreate, saveReference)
			if value, ok := gScope.Get("gorm:save_associations"); ok {
				g.log().Debug("gorm:save_associations")
				autoUpdate = checkTruth(value)
				autoCreate = autoUpdate
				g.log().Debugf("AutoUpdate: %v, autoCreate: %v", autoUpdate, autoCreate)
			} else if value, ok := field.TagSettings["SAVE_ASSOCIATIONS"]; ok {
				g.log().Debugf("SAVE_ASSOCIATIONS")
				autoUpdate = checkTruth(value)
				autoCreate = autoUpdate
			}

			if value, ok := gScope.Get("gorm:association_autoupdate"); ok {
				g.log().Debug("gorm:association_autoupdate")
				autoUpdate = checkTruth(value)
				g.log().Debug("ASSOCIATION_AUTOUPDATE")
			} else if value, ok := field.TagSettings["ASSOCIATION_AUTOUPDATE"]; ok {
				autoUpdate = checkTruth(value)
			}

			if value, ok := gScope.Get("gorm:association_autocreate"); ok {
				g.log().Debug("gorm:association_autocreate")
				autoCreate = checkTruth(value)
			} else if value, ok := field.TagSettings["ASSOCIATION_AUTOCREATE"]; ok {
				g.log().Debug("ASSOCIATION_AUTOCREATE")
				autoCreate = checkTruth(value)
			}

			if value, ok := gScope.Get("gorm:association_save_reference"); ok {
				g.log().Debug("gorm:association_save_reference")
				saveReference = checkTruth(value)
			} else if value, ok := field.TagSettings["ASSOCIATION_SAVE_REFERENCE"]; ok {
				g.log().Debug("ASSOCIATION_SAVE_REFERENCE")
				saveReference = checkTruth(value)
			}
			g.log().Debugf("AutoUpdate: %v, autoCreate: %v", autoUpdate, autoCreate)
		}
	}

	return
}

func (g *GORMRepository) saveAfterAssociationsCallback(gScope *gorm.Scope) {
	g.log().Debug("saveAfterAssociationsCallback")
	jScope, ok := g.getJScope(gScope)
	if !ok {
		g.log().Debugf("scope.Scope not found for the gScope: %#v", gScope)
	}

	if jScope != nil {
		gormType := gScope.GetModelStruct().ModelType

		if jScope.Struct().Type() != gormType {
			// g.log().Warningf("Scope type doesn't match. JScope: %v, GormScope: %#v", jScope.Struct.GetType(), gScope)
			jScope = nil
		}

	}

	for _, field := range gScope.Fields() {
		// g.log().Debugf("Field within scopes: %v", field.Name)
		autoUpdate, autoCreate, saveReference, relationship := g.saveAssociationCheck(gScope, field, jScope)
		// g.log().Infof("AutoUpdate: %v, AutoCreate: %v, saveReference: %v", autoUpdate, autoCreate, saveReference)
		if relationship != nil && (relationship.Kind == "has_one" ||
			relationship.Kind == "has_many" ||
			relationship.Kind == "many_to_many") {
			value := field.Field
			// g.log().Debugf("Relationship: %v", relationship.Kind)
			// g.log().Debugf("Field: %s", field.Name)
			// g.log().Infof("AutoUpdate: %v, AutoCreate: %v, saveReference: %v", autoUpdate, autoCreate, saveReference)

			switch value.Kind() {
			case reflect.Slice:
				for i := 0; i < value.Len(); i++ {
					newDB := gScope.NewDB()
					g.log().Debug("NewDB for field: %s", field.Name)
					elem := value.Index(i).Addr().Interface()
					newScope := newDB.NewScope(elem)

					if saveReference {
						if relationship.JoinTableHandler == nil && len(relationship.ForeignFieldNames) != 0 {
							for idx, fieldName := range relationship.ForeignFieldNames {
								associationForeignName := relationship.AssociationForeignDBNames[idx]
								if f, ok := gScope.FieldByName(associationForeignName); ok {
									gScope.Err(newScope.SetColumn(fieldName, f.Field.Interface()))
								}
							}
						}

						if relationship.PolymorphicType != "" {
							gScope.Err(newScope.SetColumn(relationship.PolymorphicType, relationship.PolymorphicValue))
						}
					}

					if newScope.PrimaryKeyZero() {
						if autoCreate {
							gScope.Err(newDB.Save(elem).Error)
						}
					} else if autoUpdate {
						g.log().Debug("Autoupdate")
						gScope.Err(newDB.Save(elem).Error)
					}

					if !gScope.New(newScope.Value).PrimaryKeyZero() && saveReference {
						if joinTableHandler := relationship.JoinTableHandler; joinTableHandler != nil {
							gScope.Err(joinTableHandler.Add(joinTableHandler, newDB, gScope.Value, newScope.Value))
						}
					}
				}
			default:
				elem := value.Addr().Interface()
				newScope := gScope.New(elem)

				if saveReference {
					g.log().Debug("SetReference")
					if len(relationship.ForeignFieldNames) != 0 {
						for idx, fieldName := range relationship.ForeignFieldNames {
							associationForeignName := relationship.AssociationForeignDBNames[idx]
							if f, ok := gScope.FieldByName(associationForeignName); ok {
								gScope.Err(newScope.SetColumn(fieldName, f.Field.Interface()))
							}
						}
					}

					if relationship.PolymorphicType != "" {
						gScope.Err(newScope.SetColumn(relationship.PolymorphicType, relationship.PolymorphicValue))
					}
				}

				if newScope.PrimaryKeyZero() {
					if autoCreate {
						g.log().Debugf("AutCreate single value for: %#v", elem)
						gScope.Err(gScope.NewDB().Save(elem).Error)
					}
				} else if autoUpdate {
					g.log().Debugf("AutoUpdate model: %#v", elem)
					gScope.Err(gScope.NewDB().Model(elem).Updates(elem).Error)
				}
			}

		} else {
			// g.log().Debugf("Don't go: %v", field.Name)
		}
	}
}

// saveBeforeAssociationsCallback
func (g *GORMRepository) saveBeforeAssociationsCallback(gScope *gorm.Scope) {
	_, ok := g.getJScope(gScope)

	g.log().Debug("saveBeforeAssociationsCallback")

	for _, field := range gScope.Fields() {
		relationship := field.Relationship

		if relationship != nil && relationship.Kind == "belongs_to" {
			g.log().Debugf("Checking field: %s", field.Name)
			if ok {
				continue
			}
			fieldValue := field.Field.Addr().Interface()
			newScope := gScope.New(fieldValue)

			if newScope.PrimaryKeyZero() {
				g.log().Debug("PK is zero")
				continue
			}

			if len(relationship.ForeignFieldNames) != 0 {
				// set value's foreign key
				for idx, fieldName := range relationship.ForeignFieldNames {
					associationForeignName := relationship.AssociationForeignDBNames[idx]
					if foreignField, ok := gScope.New(fieldValue).FieldByName(associationForeignName); ok {
						g.log().Debugf("Setting value: %v for field: %v", foreignField.Field.Interface(), fieldName)
						gScope.Err(gScope.SetColumn(fieldName, foreignField.Field.Interface()))
					}
				}
			}

		}
	}
}
