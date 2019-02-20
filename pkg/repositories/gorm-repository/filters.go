package gormrepo

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/kucjac/jsonapi/pkg/mapping"
	"github.com/kucjac/jsonapi/pkg/query/filters"
	"github.com/kucjac/jsonapi/pkg/query/scope"
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

func buildWhere(tableName, columnName string, filter *filters.FilterField,
) ([]*whereQ, error) {
	var err error
	wheres := []*whereQ{}
	for _, fv := range filter.Values() {
		if len(fv.Values) == 0 {
			log.Debugf("No values provided for the filter: %s", filter)
			return wheres, IErrNoValuesProvided
		}
		op := sqlizeOperator(fv.Operator())
		var valueMark string
		if fv.Operator() == filters.OpIn || fv.Operator() == filters.OpNotIn {

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
			if fv.Operator() == filters.OpStartsWith {
				for i, v := range fv.Values {
					strVal, ok := v.(string)
					if !ok {
						err = fmt.Errorf("Invalid value provided for the OpStartsWith filter: %v", reflect.TypeOf(v))
						return wheres, err
					}
					fv.Values[i] = strVal + "%"
				}

				// fmt.Println(fv.Values)
			} else if fv.Operator() == filters.OpContains {
				for i, v := range fv.Values {
					strVal, ok := v.(string)
					if !ok {
						err = fmt.Errorf("Invalid value provided for the OpStartsWith filter: %v", reflect.TypeOf(v))
						return wheres, err
					}
					fv.Values[i] = "%" + strVal + "%"
				}
			} else if fv.Operator() == filters.OpEndsWith {
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
func addWhere(db *gorm.DB, tableName, columnName string, filter *filters.FilterField) error {
	wheres, err := buildWhere(tableName, columnName, filter)
	if err != nil {
		return err
	}

	for _, wq := range wheres {
		*db = *db.Where(wq.Str, wq.values...)
	}
	return nil
}

func (g *GORMRepository) buildFilters(db *gorm.DB, mStruct *gorm.ModelStruct, s *scope.Scope,
) error {

	var (
		err       error
		gormField *gorm.StructField
	)

	for _, primary := range s.PrimaryFilters() {
		g.log().Debugf("Building filter for field: %s", primary.StructField().Name())
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

	// if given s uses i18n check if it contains language filter
	if s.LanguageFilter() != nil {
		// it should be primary field but it does not have to be primary
		gormField, err = getGormField(s.LanguageFilter(), mStruct, false)
		if err != nil {
			return err
		}

		if !gormField.IsIgnored {
			if err = addWhere(db, mStruct.TableName(db), gormField.DBName, s.LanguageFilter()); err != nil {
				return err
			}
		}

	} else {
		// No language filter ?
	}

	for _, fkFilter := range s.ForeignFilters() {
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

	for _, attrFilter := range s.AttributeFilters() {
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

	for _, relationFilter := range s.RelationFilters() {
		if rel := relationFilter.StructField().Relationship(); rel != nil {
			switch rel.Kind() {
			case mapping.RelHasMany, mapping.RelHasOne:
				if rel.Sync() == nil || (rel.Sync() != nil && *rel.Sync()) {
					continue
				}
			case mapping.RelMany2Many:
				if rel.Sync() != nil && *rel.Sync() {
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

			nesteds := relationFilter.NestedFilters()
			// The relationshipfilter
			if len(nesteds) != 1 {
				err = IErrBadRelationshipField
				return err
			}

			// The subfield of relationfilter must be a primary key
			if nesteds[0].StructField().FieldKind() != mapping.KindPrimary {
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

				err = addWhere(db, mStruct.TableName(db), foreignField.DBName, nesteds[0])
				if err != nil {
					return err
				}
			case associationHasMany:
				// has many can be found from different table
				// thus it must be added with included where
				relScope := db.NewScope(reflect.New(
					relationFilter.StructField().Relationship().ModelStruct().Type(),
				).Interface())

				relMStruct := relScope.GetModelStruct()
				relDB := relScope.DB()

				err = buildRelationFilters(relDB, relMStruct, nesteds[0])
				if err != nil {
					return err
				}
				// the query should be select foreign key from related table where filters for related table.

				// the wheres should already be added into relDB
				expr := relDB.Table(relMStruct.TableName(relDB)).Select(gormField.Relationship.ForeignDBNames[0]).QueryExpr()

				op := sqlizeOperator(filters.OpIn)
				valueMark := "(?)"
				columnName := mStruct.PrimaryFields[0].DBName
				q := fmt.Sprintf("\"%s\".\"%s\" %s %s", relMStruct.TableName(db), columnName, op, valueMark)

				*db = *db.Where(q, expr)

			case associationManyToMany:
				relScope := db.NewScope(reflect.New(relationFilter.StructField().Relationship().ModelStruct().Type()).Interface())

				relDB := relScope.DB()

				joinTableHandler := gormField.Relationship.JoinTableHandler
				// relatedModelFK := gormField.Relationship.AssociationForeignDBNames[0]

				relDB = relDB.Table(gormField.Relationship.JoinTableHandler.Table(relDB)).
					Select(joinTableHandler.SourceForeignKeys()[0].DBName)
				// fmt.Printf("%v", relDB)

				err = addWhere(relDB, joinTableHandler.Table(db), joinTableHandler.DestinationForeignKeys()[0].DBName, nesteds[0])
				if err != nil {
					g.log().Debugf("Error while createing Many2Many WHERE query: %v", err)
					return err
				}

				columnName := mStruct.PrimaryFields[0].DBName
				op := sqlizeOperator(filters.OpIn)
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
	db *gorm.DB, mStruct *gorm.ModelStruct, s *scope.Scope,
) ([]*whereQ, error) {

	var (
		err       error
		gormField *gorm.StructField
	)

	wqs := []*whereQ{}

	for _, primary := range s.PrimaryFilters() {
		g.log().Debugf("Building filter for field: %s", primary.StructField().Name())
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

	// if given s uses i18n check if it contains language filter

	if s.LanguageFilter() != nil {
		// it should be primary field but it does not have to be primary
		gormField, err = getGormField(s.LanguageFilter(), mStruct, false)
		if err != nil {
			return nil, err
		}

		if !gormField.IsIgnored {

			wq, err := buildWhere(mStruct.TableName(db), gormField.DBName, s.LanguageFilter())
			if err != nil {
				return nil, err
			}
			wqs = append(wqs, wq...)
		}
	}

	for _, fkFilter := range s.ForeignFilters() {
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

	for _, attrFilter := range s.AttributeFilters() {
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

	// iterate over relationship filter
	for _, relationFilter := range s.RelationFilters() {
		if rel := relationFilter.StructField().Relationship(); rel != nil && rel.Kind() == mapping.RelMany2Many {
			switch rel.Kind() {
			case mapping.RelHasMany, mapping.RelHasOne:
				if rel.Sync() == nil || (rel.Sync() != nil && *rel.Sync()) {
					continue
				}
			case mapping.RelMany2Many:
				if rel.Sync() != nil && *rel.Sync() {
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

			nesteds := relationFilter.NestedFilters()
			// The relationshipfilter
			if len(nesteds) != 1 {
				err = IErrBadRelationshipField
				return nil, err
			}

			// The subfield of relationfilter must be a primary key
			if nesteds[0].StructField().FieldKind() != mapping.KindPrimary {
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

				wq, err := buildWhere(mStruct.TableName(db), foreignField.DBName, nesteds[0])
				if err != nil {
					return nil, err
				}
				wqs = append(wqs, wq...)

			case associationHasMany:
				// has many can be found from different table
				// thus it must be added with included where
				relScope := db.NewScope(reflect.New(relationFilter.StructField().Relationship().ModelStruct().Type()).Interface())
				relMStruct := relScope.GetModelStruct()
				relDB := relScope.DB()

				err = buildRelationFilters(relDB, relMStruct, nesteds[0])
				if err != nil {
					return nil, err
				}
				// the query should be select foreign key from related table where filters for related table.

				// the wheres should already be added into relDB
				expr := relDB.Table(relMStruct.TableName(relDB)).Select(gormField.Relationship.ForeignDBNames[0]).QueryExpr()

				op := sqlizeOperator(filters.OpIn)
				valueMark := "(?)"
				columnName := mStruct.PrimaryFields[0].DBName
				q := fmt.Sprintf("\"%s\".\"%s\" %s %s", relMStruct.TableName(db), columnName, op, valueMark)
				wq := &whereQ{Str: q, values: []interface{}{expr}}
				wqs = append(wqs, wq)

			case associationManyToMany:
				relScope := db.NewScope(reflect.New(relationFilter.StructField().Relationship().ModelStruct().Type()).Interface())

				relDB := relScope.DB()

				joinTableHandler := gormField.Relationship.JoinTableHandler
				// relatedModelFK := gormField.Relationship.AssociationForeignDBNames[0]

				relDB = relDB.Table(gormField.Relationship.JoinTableHandler.Table(relDB)).
					Select(joinTableHandler.SourceForeignKeys()[0].DBName)
				// fmt.Printf("%v", relDB)

				err = addWhere(relDB, joinTableHandler.Table(db), joinTableHandler.DestinationForeignKeys()[0].DBName, nesteds[0])
				if err != nil {
					return nil, err
				}

				columnName := mStruct.PrimaryFields[0].DBName
				op := sqlizeOperator(filters.OpIn)
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
	filters ...*filters.FilterField,
) error {
	var (
		gormField *gorm.StructField
		err       error
	)

	for _, filter := range filters {
		var isPrimary bool
		// get gorm structField
		switch filter.StructField().FieldKind() {
		case mapping.KindPrimary:
			isPrimary = true

		case mapping.KindAttribute, mapping.KindRelationshipSingle, mapping.KindRelationshipMultiple:
			isPrimary = false
		default:
			err = fmt.Errorf("Unsupported jsonapi field type: '%v' for field: '%s' in model: '%v'.", filter.StructField().FieldKind(), filter.StructField().Name(), gormModel.ModelType)
			return err
		}
		gormField, err = getGormField(filter, gormModel, isPrimary)
		if err != nil {
			return err
		}

		if filter.StructField().FieldKind() == mapping.KindAttribute || filter.StructField().FieldKind() == mapping.KindPrimary {

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
	filterField *filters.FilterField,
	model *gorm.ModelStruct,
	isPrimary bool,
) (*gorm.StructField, error) {

	// fmt.Printf("Before: '%v' model: '%v' isPrim: '%v'\n", filterField.StructField, model.ModelType, isPrimary)
	if isPrimary {
		if len(model.PrimaryFields) == 1 {
			return model.PrimaryFields[0], nil
		} else {
			for _, prim := range model.PrimaryFields {
				if prim.Struct.Index[0] == filterField.StructField().ReflectField().Index[0] {
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
			if field.Struct.Index[0] == filterField.StructField().ReflectField().Index[0] {
				return field, nil
			}
		}
	}

	// fmt.Printf("filterField: '%+v'\n", filterField.GetReflectStructField())
	// fmt.Printf("ff ID:'%v'\n", filterField.GetReflectStructField().Index)

	return nil, fmt.Errorf("Invalid filtering field: '%v' not found in the gorm ModelStruct: '%v'", filterField.StructField().Name(), model.ModelType)
}

func sqlizeOperator(operator *filters.Operator) string {
	switch operator {
	case filters.OpEqual:
		return "="
	case filters.OpIn:
		return "IN"
	case filters.OpNotEqual:
		return "<>"
	case filters.OpNotIn:
		return "NOT IN"
	case filters.OpGreaterEqual:
		return ">="
	case filters.OpGreaterThan:
		return ">"
	case filters.OpLessEqual:
		return "<="
	case filters.OpLessThan:
		return "<"
	case filters.OpContains, filters.OpStartsWith, filters.OpEndsWith:
		return "LIKE"

	}
	return "="
}
