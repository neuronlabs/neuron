package gormrepo

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi"
	"github.com/pkg/errors"
	"reflect"
)

var (
	associationBelongsTo  = "belongs_to"
	associationHasOne     = "has_one"
	associationHasMany    = "has_many"
	associationManyToMany = "many_to_many"
)

var (
	IErrNoFieldFound     = errors.New("No field found for the relationship")
	IErrNoValuesProvided = errors.New("No values provided in the filter.")
)

type whereQ struct {
	Str    string
	values []interface{}
}

func buildWhere(tableName, columnName string, filter *jsonapi.FilterField,
) ([]*whereQ, error) {
	var err error
	wheres := []*whereQ{}
	for _, fv := range filter.Values {
		if len(fv.Values) == 0 {
			return wheres, IErrNoValuesProvided
		}
		op := sqlizeOperator(fv.Operator)
		var valueMark string
		if fv.Operator == jsonapi.OpIn || fv.Operator == jsonapi.OpNotIn {

			valueMark = "("

			for i := range fv.Values {
				valueMark += "?"
				if i != len(fv.Values)-1 {
					valueMark += ","
				}
			}

			valueMark += ")"

		} else {
			if len(fv.Values) > 1 {
				err = fmt.Errorf("Too many values for given operator: '%s', '%s'", fv.Values, fv.Operator)
				return wheres, err
			}
			valueMark = "?"
			if fv.Operator == jsonapi.OpStartsWith {
				for i, v := range fv.Values {
					strVal, ok := v.(string)
					if !ok {
						err = fmt.Errorf("Invalid value provided for the OpStartsWith filter: %v", reflect.TypeOf(v))
						return wheres, err
					}
					fv.Values[i] = strVal + "%"
				}

				// fmt.Println(fv.Values)
			} else if fv.Operator == jsonapi.OpContains {
				for i, v := range fv.Values {
					strVal, ok := v.(string)
					if !ok {
						err = fmt.Errorf("Invalid value provided for the OpStartsWith filter: %v", reflect.TypeOf(v))
						return wheres, err
					}
					fv.Values[i] = "%" + strVal + "%"
				}
			} else if fv.Operator == jsonapi.OpEndsWith {
				for i, v := range fv.Values {
					strVal, ok := v.(string)
					if !ok {
						err = fmt.Errorf("Invalid value provided for the OpStartsWith filter: %v", reflect.TypeOf(v))
						return wheres, err
					}
					fv.Values[i] = "%" + strVal
				}
			}
		}
		q := fmt.Sprintf("\"%s\".\"%s\" %s %s", tableName, columnName, op, valueMark)
		wQ := &whereQ{Str: q, values: fv.Values}
		wheres = append(wheres, wQ)
	}
	return wheres, nil
}

// addWhere adds the where to the scope of the db
func addWhere(db *gorm.DB, tableName, columnName string, filter *jsonapi.FilterField) error {
	wheres, err := buildWhere(tableName, columnName, filter)
	if err != nil {
		return err
	}

	for _, wq := range wheres {
		*db = *db.Where(wq.Str, wq.values...)
	}
	return nil
}

func (g *GORMRepository) buildFilters(db *gorm.DB, mStruct *gorm.ModelStruct, scope *jsonapi.Scope,
) error {

	var (
		err       error
		gormField *gorm.StructField
	)

	for _, primary := range scope.PrimaryFilters {
		g.log().Debugf("Building filter for field: %s", primary.GetFieldName())
		// fmt.Printf("Primary field: '%s'\n", primary.GetFieldName())
		gormField, err = getGormField(primary, mStruct, true)
		if err != nil {
			return err
		}
		if !gormField.IsIgnored {
			if err = addWhere(db, mStruct.TableName(db), gormField.DBName, primary); err != nil {
				return err
			}
		}

	}

	// if given scope uses i18n check if it contains language filter
	if scope.UseI18n() {
		if scope.LanguageFilters != nil {
			// it should be primary field but it does not have to be primary
			gormField, err = getGormField(scope.LanguageFilters, mStruct, false)
			if err != nil {
				return err
			}

			if !gormField.IsIgnored {
				if err = addWhere(db, mStruct.TableName(db), gormField.DBName, scope.LanguageFilters); err != nil {
					return err
				}
			}

		} else {
			// No language filter ?
		}
	}

	for _, fkFilter := range scope.ForeignKeyFilters {
		gormField, err = getGormField(fkFilter, mStruct, false)
		if err != nil {
			return errors.Wrapf(err, "getGormField fo ForeignKeyFilter: %#v failed", fkFilter)
		}

		if !gormField.IsIgnored {
			if err = addWhere(db, mStruct.TableName(db), gormField.DBName, fkFilter); err != nil {
				return errors.Wrapf(err, "AddWhere to ForeignKey filter: %#v failed.", fkFilter)
			}
		}
	}

	for _, attrFilter := range scope.AttributeFilters {
		// fmt.Printf("Attribute field: '%s'\n", attrFilter.GetFieldName())
		gormField, err = getGormField(attrFilter, mStruct, false)
		if err != nil {
			return err
		}

		if !gormField.IsIgnored {
			if err = addWhere(db, mStruct.TableName(db), gormField.DBName, attrFilter); err != nil {
				return err
			}
		}

	}

	for _, relationFilter := range scope.RelationshipFilters {
		if rel := relationFilter.GetRelationship(); rel != nil {
			switch rel.Kind {
			case jsonapi.RelHasMany, jsonapi.RelHasOne:
				if rel.Sync == nil || (rel.Sync != nil && *rel.Sync) {
					continue
				}
			case jsonapi.RelMany2Many:
				if rel.Sync != nil && *rel.Sync {
					continue
				}
			}

			gormField, err = getGormField(relationFilter, mStruct, false)
			if err != nil {
				return err
			}

			if gormField.IsIgnored {
				continue
			}

			// The relationshipfilter
			if len(relationFilter.Relationships) != 1 {
				err = IErrBadRelationshipField
				return err
			}

			// The subfield of relationfilter must be a primary key
			if !relationFilter.Relationships[0].IsPrimary() {
				err = IErrBadRelationshipField
				return err
			}

			switch gormField.Relationship.Kind {
			case associationBelongsTo, associationHasOne:

				// BelongsTo and HasOne relationship should contain foreign field in the same struct
				// The foreign field should contain foreign key
				foreignFieldName := gormField.Relationship.ForeignFieldNames[0]
				var found bool
				var foreignField *gorm.StructField

				// find the field in gorm model struct
				for _, field := range mStruct.StructFields {
					if field.Name == foreignFieldName {
						found = true
						foreignField = field
						break
					}
				}

				// check fi field was found
				if !found {
					err = IErrNoFieldFound
					return err
				}

				err = addWhere(db, mStruct.TableName(db), foreignField.DBName, relationFilter.Relationships[0])
				if err != nil {
					return err
				}
			case associationHasMany:
				// has many can be found from different table
				// thus it must be added with included where
				relScope := db.NewScope(reflect.New(relationFilter.GetRelatedModelType()).Interface())
				relMStruct := relScope.GetModelStruct()
				relDB := relScope.DB()

				err = buildRelationFilters(relDB, relMStruct, relationFilter.Relationships[0])
				if err != nil {
					return err
				}
				// the query should be select foreign key from related table where filters for related table.

				// the wheres should already be added into relDB
				expr := relDB.Table(relMStruct.TableName(relDB)).Select(gormField.Relationship.ForeignDBNames[0]).QueryExpr()

				op := sqlizeOperator(jsonapi.OpIn)
				valueMark := "(?)"
				columnName := mStruct.PrimaryFields[0].DBName
				q := fmt.Sprintf("\"%s\".\"%s\" %s %s", relMStruct.TableName(db), columnName, op, valueMark)

				*db = *db.Where(q, expr)

			case associationManyToMany:
				relScope := db.NewScope(reflect.New(relationFilter.GetRelatedModelType()).Interface())

				relDB := relScope.DB()

				joinTableHandler := gormField.Relationship.JoinTableHandler
				// relatedModelFK := gormField.Relationship.AssociationForeignDBNames[0]

				relDB = relDB.Table(gormField.Relationship.JoinTableHandler.Table(relDB)).
					Select(joinTableHandler.SourceForeignKeys()[0].DBName)
				// fmt.Printf("%v", relDB)

				err = addWhere(relDB, joinTableHandler.Table(db), joinTableHandler.DestinationForeignKeys()[0].DBName, relationFilter.Relationships[0])
				if err != nil {
					g.log().Debugf("Error while createing Many2Many WHERE query: %v", err)
					return err
				}

				columnName := mStruct.PrimaryFields[0].DBName
				op := sqlizeOperator(jsonapi.OpIn)
				valueMark := "(?)"
				q := fmt.Sprintf("%s %s %s", columnName, op, valueMark)

				g.log().Debug("Many2Many filter query: %s", q)

				*db = *db.Where(q, relDB.QueryExpr())

				// err= buildRelationFilters(relDB, relMStruct, ...)
			}
		}
	}

	return nil
}

func (g *GORMRepository) getQueryFilters(
	db *gorm.DB, mStruct *gorm.ModelStruct, scope *jsonapi.Scope,
) ([]*whereQ, error) {

	var (
		err       error
		gormField *gorm.StructField
	)

	wqs := []*whereQ{}

	for _, primary := range scope.PrimaryFilters {
		g.log().Debugf("Building filter for field: %s", primary.GetFieldName())
		// fmt.Printf("Primary field: '%s'\n", primary.GetFieldName())
		gormField, err = getGormField(primary, mStruct, true)
		if err != nil {
			return nil, err
		}
		if !gormField.IsIgnored {

			wq, err := buildWhere(mStruct.TableName(db), gormField.DBName, primary)
			if err != nil {
				return nil, err
			}
			wqs = append(wqs, wq...)
		}

	}

	// if given scope uses i18n check if it contains language filter
	if scope.UseI18n() {
		if scope.LanguageFilters != nil {
			// it should be primary field but it does not have to be primary
			gormField, err = getGormField(scope.LanguageFilters, mStruct, false)
			if err != nil {
				return nil, err
			}

			if !gormField.IsIgnored {

				wq, err := buildWhere(mStruct.TableName(db), gormField.DBName, scope.LanguageFilters)
				if err != nil {
					return nil, err
				}
				wqs = append(wqs, wq...)
			}

		} else {
			// No language filter ?
		}
	}

	for _, fkFilter := range scope.ForeignKeyFilters {
		gormField, err = getGormField(fkFilter, mStruct, false)
		if err != nil {
			return nil, errors.Wrapf(err, "getGormField fo ForeignKeyFilter: %#v failed", fkFilter)
		}

		if !gormField.IsIgnored {

			wq, err := buildWhere(mStruct.TableName(db), gormField.DBName, fkFilter)
			if err != nil {
				return nil, err
			}
			wqs = append(wqs, wq...)
		}
	}

	for _, attrFilter := range scope.AttributeFilters {
		// fmt.Printf("Attribute field: '%s'\n", attrFilter.GetFieldName())
		gormField, err = getGormField(attrFilter, mStruct, false)
		if err != nil {
			return nil, err
		}

		if !gormField.IsIgnored {

			wq, err := buildWhere(mStruct.TableName(db), gormField.DBName, attrFilter)
			if err != nil {
				return nil, err
			}
			wqs = append(wqs, wq...)
		}

	}

	for _, relationFilter := range scope.RelationshipFilters {
		if rel := relationFilter.GetRelationship(); rel != nil && rel.IsManyToMany() {
			switch rel.Kind {
			case jsonapi.RelHasMany, jsonapi.RelHasOne:
				if rel.Sync == nil || (rel.Sync != nil && *rel.Sync) {
					continue
				}
			case jsonapi.RelMany2Many:
				if rel.Sync != nil && *rel.Sync {
					continue
				}
			}

			gormField, err = getGormField(relationFilter, mStruct, false)
			if err != nil {
				return nil, err
			}

			if gormField.IsIgnored {
				continue
			}

			// The relationshipfilter
			if len(relationFilter.Relationships) != 1 {
				err = IErrBadRelationshipField
				return nil, err
			}

			// The subfield of relationfilter must be a primary key
			if !relationFilter.Relationships[0].IsPrimary() {
				err = IErrBadRelationshipField
				return nil, err
			}

			switch gormField.Relationship.Kind {
			case associationBelongsTo, associationHasOne:

				// BelongsTo and HasOne relationship should contain foreign field in the same struct
				// The foreign field should contain foreign key
				foreignFieldName := gormField.Relationship.ForeignFieldNames[0]
				var found bool
				var foreignField *gorm.StructField

				// find the field in gorm model struct
				for _, field := range mStruct.StructFields {
					if field.Name == foreignFieldName {
						found = true
						foreignField = field
						break
					}
				}

				// check fi field was found
				if !found {
					err = IErrNoFieldFound
					return nil, err
				}

				wq, err := buildWhere(mStruct.TableName(db), foreignField.DBName, relationFilter.Relationships[0])
				if err != nil {
					return nil, err
				}
				wqs = append(wqs, wq...)

			case associationHasMany:
				// has many can be found from different table
				// thus it must be added with included where
				relScope := db.NewScope(reflect.New(relationFilter.GetRelatedModelType()).Interface())
				relMStruct := relScope.GetModelStruct()
				relDB := relScope.DB()

				err = buildRelationFilters(relDB, relMStruct, relationFilter.Relationships[0])
				if err != nil {
					return nil, err
				}
				// the query should be select foreign key from related table where filters for related table.

				// the wheres should already be added into relDB
				expr := relDB.Table(relMStruct.TableName(relDB)).Select(gormField.Relationship.ForeignDBNames[0]).QueryExpr()

				op := sqlizeOperator(jsonapi.OpIn)
				valueMark := "(?)"
				columnName := mStruct.PrimaryFields[0].DBName
				q := fmt.Sprintf("\"%s\".\"%s\" %s %s", relMStruct.TableName(db), columnName, op, valueMark)
				wq := &whereQ{Str: q, values: []interface{}{expr}}
				wqs = append(wqs, wq)

			case associationManyToMany:
				relScope := db.NewScope(reflect.New(relationFilter.GetRelatedModelType()).Interface())

				relDB := relScope.DB()

				joinTableHandler := gormField.Relationship.JoinTableHandler
				// relatedModelFK := gormField.Relationship.AssociationForeignDBNames[0]

				relDB = relDB.Table(gormField.Relationship.JoinTableHandler.Table(relDB)).
					Select(joinTableHandler.SourceForeignKeys()[0].DBName)
				// fmt.Printf("%v", relDB)

				err = addWhere(relDB, joinTableHandler.Table(db), joinTableHandler.DestinationForeignKeys()[0].DBName, relationFilter.Relationships[0])
				if err != nil {
					return nil, err
				}

				columnName := mStruct.PrimaryFields[0].DBName
				op := sqlizeOperator(jsonapi.OpIn)
				valueMark := "(?)"
				q := fmt.Sprintf("%s %s %s", columnName, op, valueMark)

				wqs = append(wqs, &whereQ{q, []interface{}{relDB.QueryExpr()}})
			}
		}
	}

	return wqs, nil
}

func buildRelationFilters(
	db *gorm.DB,
	gormModel *gorm.ModelStruct,
	filters ...*jsonapi.FilterField,
) error {
	var (
		gormField *gorm.StructField
		err       error
	)

	for _, filter := range filters {
		var isPrimary bool
		// get gorm structField
		switch filter.GetFieldKind() {
		case jsonapi.Primary:
			isPrimary = true
		case jsonapi.Attribute, jsonapi.RelationshipSingle, jsonapi.RelationshipMultiple:
			isPrimary = false
		default:
			err = fmt.Errorf("Unsupported jsonapi field type: '%v' for field: '%s' in model: '%v'.", filter.GetFieldKind(), filter.GetFieldName(), gormModel.ModelType)
			return err
		}
		gormField, err = getGormField(filter, gormModel, isPrimary)
		if err != nil {
			return err
		}

		if filter.GetFieldKind() == jsonapi.Attribute || filter.GetFieldKind() == jsonapi.Primary {

			err = addWhere(db, gormModel.TableName(db), gormField.DBName, filter)
			if err != nil {
				return err
			}
		} else {
			// no direct getter for table name
			err = IErrBadRelationshipField
			return err
			// relScope := db.NewScope(reflect.New(filter.GetRelatedModelType()).Interface())
			// relMStruct := relScope.GetModelStruct()
			// relDB := relScope.DB()
			// err = buildRelationFilters(relDB, relMStruct, filter.Relationships...)
			// if err != nil {
			// 	return err
			// }
			// expr := relDB.Table(relMStruct.TableName(relDB)).Select(relScope.PrimaryField().DBName).QueryExpr()
			// *db = *db.Where(gormField.DBName, expr)
		}
	}
	return nil
}

func getGormField(
	filterField *jsonapi.FilterField,
	model *gorm.ModelStruct,
	isPrimary bool,
) (*gorm.StructField, error) {

	// fmt.Printf("Before: '%v' model: '%v' isPrim: '%v'\n", filterField.StructField, model.ModelType, isPrimary)
	if isPrimary {
		if len(model.PrimaryFields) == 1 {
			return model.PrimaryFields[0], nil
		} else {
			for _, prim := range model.PrimaryFields {
				if prim.Struct.Index[0] == filterField.GetFieldIndex() {
					return prim, nil
				}
			}
		}
		// } else {
		// 	// fmt.Println("Powinno wejść o tutaj.")
		// 	model.PrimaryFields
		// 	return model.PrimaryFields[0], nil
		// }
	} else {
		for _, field := range model.StructFields {
			if field.Struct.Index[0] == filterField.GetFieldIndex() {
				return field, nil
			}
		}
	}

	// fmt.Printf("filterField: '%+v'\n", filterField.GetReflectStructField())
	// fmt.Printf("ff ID:'%v'\n", filterField.GetReflectStructField().Index)

	return nil, fmt.Errorf("Invalid filtering field: '%v' not found in the gorm ModelStruct: '%v'", filterField.GetFieldName(), model.ModelType)
}

func sqlizeOperator(operator jsonapi.FilterOperator) string {
	switch operator {
	case jsonapi.OpEqual:
		return "="
	case jsonapi.OpIn:
		return "IN"
	case jsonapi.OpNotEqual:
		return "<>"
	case jsonapi.OpNotIn:
		return "NOT IN"
	case jsonapi.OpGreaterEqual:
		return ">="
	case jsonapi.OpGreaterThan:
		return ">"
	case jsonapi.OpLessEqual:
		return "<="
	case jsonapi.OpLessThan:
		return "<"
	case jsonapi.OpContains, jsonapi.OpStartsWith, jsonapi.OpEndsWith:
		return "LIKE"

	}
	return "="
}
