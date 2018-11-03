package gormrepo

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi"
	"github.com/kucjac/jsonapi/repositories"
	"github.com/kucjac/uni-db"
	"github.com/pkg/errors"
	"reflect"
)

func (g *GORMRepository) Patch(scope *jsonapi.Scope) error {
	g.log().Debug("START PATCH")
	defer func() {
		g.log().Debug("FINISHED PATCH")
	}()
	db := g.db.New()
	db = db.Begin()

	/**

	  PATCH: HANDLE NIL VALUE

	*/
	if scope.Value == nil {
		// if no value then error
		dbErr := unidb.ErrInternalError.New()
		dbErr.Message = "No value for patch method."
		return dbErr
	}

	/**

	  PATCH: PREPARE GORM SCOPE

	*/
	g.setJScope(scope, db)

	modelStruct := db.NewScope(scope.Value).GetModelStruct()

	if err := g.buildFilters(db, modelStruct, scope); err != nil {
		g.log().Debugf("BuildingFilters failed: %v", err)
		db.Rollback()
		g.log().Debugf("Rollback err: %v", db.Error)
		return g.converter.Convert(err)
	}

	/**

	  PATCH: HOOK BEFORE PATCH

	*/
	if beforePatcher, ok := scope.Value.(repositories.HookRepoBeforePatch); ok {
		if err := beforePatcher.RepoBeforePatch(db, scope); err != nil {
			g.log().Debugf("RepoBeforePatch failed. %v", err)
			db.Rollback()
			g.log().Debugf("Rollback err: %v", db.Error)
			return g.converter.Convert(err)
		}
	}

	/**

	  PATCH: UPDATE RECORD WITIHN DATABASE

	*/

	// fields := getUpdatedGormFieldNames(modelStruct, scope)

	// fieldNames := g.getSelectedGormFieldValues(modelStruct, scope.SelectedFields...)
	values := g.getUpdatedFieldValues(modelStruct, scope)

	if len(values) > 0 {
		db = db.Table(modelStruct.TableName(db)).Updates(values)
		if err := db.Error; err != nil {
			g.log().Errorf("GormRepo Update failed. %v", err)
			db.Rollback()
			g.log().Debugf("Rollback err: %v", db.Error)
			return g.converter.Convert(err)
		}
		g.log().Debugf("Updated correctly.")
	}
	err := g.patchNonSyncedRelations(scope, modelStruct, db)
	if err != nil {
		db.Rollback()
		g.log().Debugf("Rollback err: %v", db.Error)
		return g.converter.Convert(err)
	}

	/**

	  PATCH: HOOK AFTER PATCH

	*/
	if afterPatcher, ok := scope.Value.(repositories.HookRepoAfterPatch); ok {
		if err := afterPatcher.RepoAfterPatch(db, scope); err != nil {
			db.Rollback()
			return g.converter.Convert(err)
		}
	}

	db.Commit()
	return nil
}

func (g *GORMRepository) patchNonSyncedRelations(
	scope *jsonapi.Scope,
	mStruct *gorm.ModelStruct,
	rootDB *gorm.DB,
) (err error) {
	// GET PRIMARIES
	db := rootDB

	primaries := []interface{}{}
	var primsTaken bool

	getPrimaries := func() error {
		primsTaken = true
		wqs, err := g.getQueryFilters(db, mStruct, scope)
		var wheres string
		values := []interface{}{}
		for _, wq := range wqs {
			wheres += wq.Str + " AND "
			values = append(values, wq.values...)
		}

		if len(wheres) > 0 {
			wheres = wheres[:len(wheres)-5]
		} else {
			return nil
		}

		primaryFieldNames := g.getSelectedGormFieldValues(mStruct, scope.Struct.GetPrimaryField())
		q := fmt.Sprintf("SELECT \"%s\" FROM \"%s\" WHERE (%s)", primaryFieldNames[0], mStruct.TableName(rootDB), wheres)

		g.log().Debugf("Query: %s", q)

		rows, err := db.New().Raw(q, values...).Rows()
		// if err := db..Select(primaryFieldNames[0]).Find(&primaries).Error; err != nil {
		if err != nil {
			g.log().Errorf("Error while getting rows: %v", err)
			return errors.Wrapf(err, "Getting primaries for model: %s failed.", mStruct.ModelType.Name())
		}

		err = func() error {
			defer rows.Close()
			for rows.Next() {
				cols, _ := rows.Columns()
				g.log().Debugf("Columns: %v", cols)
				colTypes, _ := rows.ColumnTypes()
				var tps string
				for _, tp := range colTypes {
					tps += tp.Name() + ", "
				}
				g.log().Debugf("ColumnTypes: %+v", tps)
				switch scope.Struct.GetPrimaryField().GetReflectStructField().Type.Kind() {
				case reflect.Int:
					var i int
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.Int8:
					var i int8
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.Int16:
					var i int16
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.Int32:
					var i int32
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.Int64:
					var i int64
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.Uint:
					var i uint
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.Uint8:
					var i uint8
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.Uint16:
					var i uint16
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.Uint32:
					var i uint32
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.Uint64:
					var i uint64
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				case reflect.String:
					var i string
					err := rows.Scan(&i)
					if err != nil {
						return err
					}
					primaries = append(primaries, i)
				default:
					return errors.Errorf("Unknown primary field type: %v", scope.Struct.GetPrimaryField().GetReflectStructField().Type)
				}
			}
			return nil
		}()
		if err != nil {
			return errors.Wrap(err, "Rows read failed.")
		}
		g.log().Debugf("Primaries found:%v", primaries)
		return nil
	}

	for _, field := range scope.SelectedFields {
		g.log().Debugf("Field: %v", field.GetFieldName())
		if field.IsRelationship() {

			rel := field.GetRelationship()
			switch rel.Kind {
			case jsonapi.RelBelongsTo:
				continue
			case jsonapi.RelHasOne:
				// if has one is non synced get the value
				if rel.Sync != nil && !*rel.Sync {
					if !primsTaken {
						err := getPrimaries()
						if err != nil {
							return err
						}
					}
					if len(primaries) == 0 {
						g.log().Debugf("Relation HasOne NonSynced:'%s' for Model: '%s' not patched. No matched primaries found.", field.GetFieldName(), scope.Struct.GetType().Name())
						return nil
					}

					if len(primaries) > 1 {
						return errors.Errorf("Invalid update relation operation for model: %s relation: %s. Too many primary filter values for HasOne relationship", scope.Struct.GetType().Name(), field.GetFieldName())
					}
					// get primary value
					v := reflect.ValueOf(scope.Value)
					if v.Kind() == reflect.Ptr {
						v = v.Elem()
					}
					vField := v.FieldByIndex(field.GetReflectStructField().Index)
					primVal := reflect.ValueOf(primaries[0])

					relScope := db.NewScope(reflect.New(vField.Type()).Interface())
					g.log().Debugf("Rel TableName: %v", relScope.GetModelStruct().TableName(db))
					relPrim := field.GetRelatedModelStruct().GetPrimaryField()

					// Get GORM DBNames for the relation.id and relation.foreign fields
					var gPrimField, gForeignField *gorm.StructField
					for _, gField := range relScope.GetModelStruct().StructFields {
						g.log().Debugf("Gorm Field: %s", gField.Name)
						if gPrimField == nil {
							if isFieldEqual(gField, relPrim) {
								gPrimField = gField
								g.log().Debugf("RelPrim found: %s", gField.Name)
							}
						}
						if gForeignField == nil {
							if isFieldEqual(gField, rel.ForeignKey) {
								gForeignField = gField
								g.log().Debugf("FkField found: %s", gField.Name)
							}
						}
						if gPrimField != nil && gForeignField != nil {
							g.log().Debug("Both are found")
							break
						}
					}

					// Both Primary and Foreign should not be nil
					if gPrimField == nil {
						return errors.Errorf("Primary Key Field: %s for the relation: '%s' in model: %s not found within the model scope.", relPrim.GetFieldName(), field.GetFieldName(), relScope.GetModelStruct().ModelType.Name())
					}
					if gForeignField == nil {
						return errors.Errorf("Foreign Key Field: %s for the relation: '%s' in model: %s not found within the model scope.", rel.ForeignKey.GetFieldName(), field.GetFieldName(), relScope.GetModelStruct().ModelType.Name())
					}

					// If FieldValue is not nil set the values of the foreign key to the
					// root.primary
					if !vField.IsNil() {
						if vField.Kind() == reflect.Ptr {
							vField = vField.Elem()
						}

						relPrimVal := vField.FieldByIndex(relPrim.GetReflectStructField().Index)

						setForeignSQL := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?",
							relScope.GetModelStruct().TableName(db),
							relScope.Quote(gForeignField.DBName),
							relScope.Quote(relScope.GetModelStruct().TableName(db))+"."+relScope.Quote(gPrimField.DBName),
						)

						primaryValue := primVal.Interface()
						if reflect.DeepEqual(primVal.Interface(), reflect.Zero(primVal.Type()).Interface()) {
							primaryValue = nil
						}

						err := db.Exec(setForeignSQL, primaryValue, relPrimVal.Interface()).Error
						if err != nil {
							errors.Wrapf(err, "Update HasOne NonSynced relationship failed. Model: %s, Relationship: %s", scope.Struct.GetType(), field.GetFieldName())
						}
					} else {
						// If fieldValue is nil erease the relationship (set the foreign key to
						// NULL) for:
						// UPDATE relation.table SET foreign = NULL WHERE relation.table.id IN (SELECT id FROM relation.table WHERE foreign = root.primary)

						clearSQL := fmt.Sprintf("UPDATE %s SET %s = NULL WHERE %s IN ( SELECT %s FROM %s WHERE %s = ?)",
							relScope.Quote(relScope.GetModelStruct().TableName(db)),
							relScope.Quote(gForeignField.DBName),
							relScope.Quote(relScope.GetModelStruct().TableName(db))+"."+relScope.Quote(gPrimField.DBName),
							relScope.Quote(gPrimField.DBName),
							relScope.Quote(relScope.GetModelStruct().TableName(db)),
							relScope.Quote(relScope.GetModelStruct().TableName(db))+"."+relScope.Quote(gForeignField.DBName),
						)
						if err = relScope.DB().Exec(clearSQL, primaries[0]).Error; err != nil {
							g.log().Debugf("ClearSQL Error for relationship has one: %v", err)
							dbErr := g.converter.Convert(err)
							dbErr.Message = err.Error()
							if !dbErr.Compare(unidb.ErrNoResult) {
								return dbErr
							}
						}
					}
				} else {
					continue
				}
			case jsonapi.RelHasMany:
				if rel.Sync != nil && !*rel.Sync {
					if !primsTaken {
						err := getPrimaries()
						if err != nil {
							return err
						}
					}
					if len(primaries) == 0 {
						g.log().Debugf("Relation HasMany NonSynced:'%s' for Model: '%s' not patched. No matched primaries found.", field.GetFieldName(), scope.Struct.GetType().Name())
						return nil
					}
					// if len(primaries) > 1 {
					// 	return errors.Errorf("Invalid update relation operation for model: %s relation: %s. Too many primary filter values for HasOne relationship", scope.Struct.GetType().Name(), field.GetFieldName())
					// }
					// get primary value
					v := reflect.ValueOf(scope.Value)
					if v.Kind() == reflect.Ptr {
						v = v.Elem()
					}
					vField := v.FieldByIndex(field.GetReflectStructField().Index)
					if vField.Kind() != reflect.Slice {
						return errors.Errorf("Invalid HasMany field value. Model: %s, Field: %s. Type: %v", scope.Struct.GetType().Name(), field.GetFieldName(), field.GetFieldType().Name())
					}

					relScope := g.db.NewScope(reflect.New(vField.Type().Elem().Elem()).Interface())
					relPrim := field.GetRelatedModelStruct().GetPrimaryField()
					var primField, fkField *gorm.StructField

					for _, gField := range relScope.GetModelStruct().StructFields {
						if isFieldEqual(gField, rel.ForeignKey) {
							fkField = gField
							if primField != nil && fkField != nil {
								break
							}
							continue
						} else if isFieldEqual(gField, relPrim) {
							primField = gField
							if primField != nil && fkField != nil {
								break
							}
							continue
						}
					}

					if primField == nil {
						return errors.Errorf("No primary field found for model: %s", relScope.GetModelStruct().ModelType.Name())
					}

					if fkField == nil {
						return errors.Errorf("No foreign key field: '%s' found for model: %s", rel.ForeignKey.GetFieldName(), relScope.GetModelStruct().ModelType.Name())
					}

					if vField.Len() != 0 {
						g.log().Debugf("Relation field greater contains more than 0 entries.")

						// Get Relation Primary values
						relPrimValues := []interface{}{}

						// Before updating any relation table clear all associated entries
						// the query should look something like this
						//
						// not in relPrimValues
						//
						// UPDATE relation.table SET foreign_key = NULL WHERE relation.table.id IN
						// (SELECT id FROM relation.table WHERE foreign_key = root.primary && id)

						// When the relations are clear set the new associated entries into
						// database
						// UPDATE relation.table SET foreign_key = root.id WHERE relation.table.id // IN 'relPrimValues)'

						for i := 0; i < vField.Len(); i++ {
							elem := vField.Index(i)
							if elem.IsNil() {
								continue
							}
							if elem.Kind() == reflect.Ptr {
								elem = elem.Elem()
							}

							relPrimVal := elem.FieldByIndex(relPrim.GetReflectStructField().Index)
							relPrimValue := relPrimVal.Interface()

							// if the primary value is zero continue to next
							if reflect.DeepEqual(relPrimValue, reflect.Zero(relPrimVal.Type()).Interface()) {
								continue
							}
							relPrimValues = append(relPrimValues, relPrimValue)
						}

						// If all the primary values were Zero continue
						if len(relPrimValues) == 0 {
							g.log().Debug("relPrimValues are zero")
							continue
						}

						var relPrimQuotationMarks string

						for range relPrimValues {
							relPrimQuotationMarks += "?,"
						}

						if len(relPrimValues) > 0 {
							relPrimQuotationMarks = relPrimQuotationMarks[:len(relPrimQuotationMarks)-1]
						}

						clearSQL := fmt.Sprintf(`UPDATE %s SET %s = ? WHERE %s IN ( SELECT %s FROM %s WHERE %s = ? AND %s NOT IN (%s))`,
							relScope.Quote(relScope.GetModelStruct().TableName(db)),
							relScope.Quote(fkField.DBName),
							relScope.Quote(relScope.GetModelStruct().TableName(db))+"."+relScope.Quote(primField.DBName),
							relScope.Quote(primField.DBName),
							relScope.Quote(relScope.GetModelStruct().TableName(db)),
							relScope.Quote(relScope.GetModelStruct().TableName(db))+"."+relScope.Quote(fkField.DBName),
							relScope.Quote(relScope.GetModelStruct().TableName(db))+"."+relScope.Quote(primField.DBName),
							relPrimQuotationMarks,
						)
						g.log().Debugf("ClearSQL for  relationship HasMany: %s", clearSQL)
						clearValues := []interface{}{nil, primaries[0]}
						clearValues = append(clearValues, relPrimValues...)

						err = db.Exec(clearSQL, clearValues...).Error

						// err = relScope.DB().Table(relScope.TableName()).Updates(fieldValues).Error
						if err != nil {
							errors.Wrapf(err, "Update HasOne NonSynced relationship failed. Model: %s, Relationship: %s", scope.Struct.GetType(), field.GetFieldName())
						}

						updateSQL := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s IN (%s)",
							relScope.Quote(relScope.GetModelStruct().TableName(db)),
							relScope.Quote(fkField.DBName),
							relScope.Quote(relScope.GetModelStruct().TableName(db))+"."+relScope.Quote(primField.DBName),
							relPrimQuotationMarks,
						)

						g.log().Debugf("UpdateSQL for HasMany model %s", updateSQL)
						updateValues := append([]interface{}{primaries[0]}, relPrimValues...)
						err = db.Exec(updateSQL, updateValues...).Error
						if err != nil {
							errors.Wrapf(err, "Update HasOne NonSynced relationship failed. Model: %s, Relationship: %s", scope.Struct.GetType(), field.GetFieldName())
						}

					} else {
						clearSQL := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s IN (SELECT %s FROM %s WHERE %s = ?)",
							relScope.QuotedTableName(),
							relScope.Quote(fkField.DBName),
							relScope.QuotedTableName()+"."+relScope.Quote(primField.DBName),
							relScope.Quote(primField.DBName),
							relScope.QuotedTableName(),
							relScope.QuotedTableName()+"."+relScope.Quote(fkField.DBName),
						)
						g.log().Debugf("ClearSQL for relation HasMany: %s", clearSQL)
						clearValues := []interface{}{nil, primaries[0]}
						if err := db.Exec(clearSQL, clearValues...).Error; err != nil {
							g.log().Debugf("ClearValues failed for relation HasMany: %v", err)
							return errors.Wrapf(err, "Clearing relation values failed for model: %s relation: %s.", scope.Struct.GetType().Name(), field.GetFieldName())
						}
					}
				} else {
					continue
				}
			case jsonapi.RelMany2Many:
				if rel.Sync != nil && *rel.Sync {
					continue
				}
				if !primsTaken {
					err := getPrimaries()
					if err != nil {
						return err
					}
				}

				if len(primaries) == 0 {
					g.log().Debugf("No primary matched")
					return nil
				}

				// Prepare quotation makrs for all primary values
				var primaryQuotationMarks string
				for range primaries {
					primaryQuotationMarks += "?,"
				}

				if len(primaries) > 0 {
					primaryQuotationMarks = primaryQuotationMarks[:len(primaryQuotationMarks)-1]
				}

				relScope := db.NewScope(reflect.New(field.GetFieldType()).Interface())

				var gRelationField *gorm.StructField
				for _, gField := range mStruct.StructFields {
					if isFieldEqual(gField, field) {
						gRelationField = gField
						break
					}
				}

				if gRelationField == nil {
					return errors.Errorf("Relation field '%s' not found within gorm.Structure for model: '%s'", field.GetFieldName(), field.GetRelatedModelStruct().GetType().Name())
				}

				if gRelationField.Relationship == nil {
					return errors.Errorf("GormRelation field: %s does not contain relationship struct for model: '%s'", field.GetFieldName(), field.GetRelatedModelStruct().GetType().Name())
				}

				if gRelationField.Relationship.Kind != "many_to_many" {
					return errors.Errorf("GORM Relationship for field: '%s' is not of many2many type. Model: '%s'", field.GetFieldName(), field.GetRelatedModelStruct().GetType().Name())
				}

				// DELETE FROM jointable WHERE associated_root.id IN ( primaries)
				clearRelations := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)",
					relScope.Quote(gRelationField.Relationship.JoinTableHandler.Table(db)),
					relScope.Quote(gRelationField.Relationship.ForeignDBNames[0]),
					primaryQuotationMarks,
				)

				g.log().Debugf("Clear Relations SQL for relation many2many: %s", clearRelations)

				err = relScope.DB().Exec(clearRelations, primaries...).Error
				if err != nil {
					g.log().Debugf("ClearRelations failed for relationship many2many. %v", err)
					dbErr := g.converter.Convert(err)
					dbErr.Message = err.Error()
					if !dbErr.Compare(unidb.ErrNoResult) {
						return dbErr
					}
				}

				v := reflect.ValueOf(scope.Value)
				if v.Kind() == reflect.Ptr {
					v = v.Elem()
				}

				vField := v.FieldByIndex(field.GetReflectStructField().Index)
				insertSQL := fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES(?, ?)",
					relScope.Quote(gRelationField.Relationship.JoinTableHandler.Table(db)),
					relScope.Quote(gRelationField.Relationship.AssociationForeignDBNames[0]),
					relScope.Quote(gRelationField.Relationship.ForeignDBNames[0]),
				)

				g.log().Debugf("SQL Inserting new many2many relationships: %s", insertSQL)
				relPrim := field.GetRelatedModelStruct().GetPrimaryField()
				for _, primary := range primaries {
					for i := 0; i < vField.Len(); i++ {
						elem := vField.Index(i)
						if elem.IsNil() {
							continue
						}

						if elem.Kind() == reflect.Ptr {
							elem = elem.Elem()
						}

						elemPrimFieldValue := elem.FieldByIndex(relPrim.GetReflectStructField().Index)
						elemPrimValue := elemPrimFieldValue.Interface()

						if reflect.DeepEqual(elemPrimValue, reflect.Zero(relPrim.GetFieldType()).Interface()) {
							continue
						}
						err = relScope.DB().Exec(insertSQL, elemPrimValue, primary).Error
						if err != nil {
							g.log().Debugf("Inserting new many2many relationships failed. %v", err)
							dbErr := g.converter.Convert(err)
							dbErr.Message = err.Error()
							return dbErr
						}
					}
				}

			}

		}
	}
	return nil
}

/**

UPDATE relation.table SET foreignkey = root.primary WHERE relation.id = relation.ID


WHILE UPDATING model with relation of type has one the primary field filter contain more than one value

for both HasMany and HasOne

PRIMFILTER(1,2,3) RelationsID(4,5,6)

for i := range prims {
	UPDATE relation.table SET foreignkey = i WHERE relation.table.id IN RelationsID
}



MANY2MANY

PRIMFILTER(1,2,3) RelationsID(4,5,6)

for i := range prims {
	DELETE ALL ENTRIES CONTAINING i
	for j := range relations {
		APPEND ENTRIES i, j
	}
}


*/
