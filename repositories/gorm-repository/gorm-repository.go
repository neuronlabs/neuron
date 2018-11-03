package gormrepo

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kucjac/jsonapi"
	"github.com/kucjac/uni-db"
	"github.com/kucjac/uni-db/gormconv"
	"github.com/kucjac/uni-logger"
	"log"
	"os"
	"reflect"
	debugStack "runtime/debug"
)

var (
	idCounter uint64
)

const (
	annotationBelongsTo  = "belongs_to"
	annotationHasOne     = "has_one"
	annotationManyToMany = "many_to_many"
	annotationHasMany    = "has_many"
)

var (
	IErrBadRelationshipField = errors.New("This repository does not allow relationship filter of field different than primary.")
)

type GORMRepository struct {
	db        *gorm.DB
	converter *gormconv.GORMConverter

	ptrSize int

	logger   unilogger.LeveledLogger
	logLevel unilogger.Level
}

func New(db *gorm.DB) (*GORMRepository, error) {
	gormRepo := &GORMRepository{}
	err := gormRepo.initialize(db.New())
	if err != nil {
		return nil, err
	}

	return gormRepo, nil
}

func (g *GORMRepository) NewDB() *gorm.DB {
	db := g.db.New()
	db = db.Set("gorm:association_autoupdate", false)
	db = db.Set("gorm:association_autocreate", false)
	db = db.Set("gorm:association_save_reference", false)
	db = db.Set("gorm:save_associations", false)
	return db
}

// GetLogLevel gets the current log level.
func (g *GORMRepository) GetLogLevel() unilogger.Level {
	return g.logLevel
}

// SetLogLevel sets the log level for given unilogger.Level
func (g *GORMRepository) SetLogLevel(level unilogger.Level) {
	g.logLevel = level

	if levelSetter, ok := g.logger.(unilogger.LevelSetter); ok {
		levelSetter.SetLevel(level)
	}
}

func (g *GORMRepository) SetLogger(logger unilogger.LeveledLogger) {
	g.logger = logger
}

func (g *GORMRepository) initialize(db *gorm.DB) (err error) {
	if db == nil {
		err = errors.New("Nil pointer as an argument provided.")
		return
	}

	g.db = db

	db.Callback().Create().Replace("gorm:save_after_associations", g.saveAfterAssociationsCallback)
	db.Callback().Update().Replace("gorm:save_after_associations", g.saveAfterAssociationsCallback)

	db.Callback().Create().Replace("gorm:save_before_associations", g.saveBeforeAssociationsCallback)
	db.Callback().Update().Replace("gorm:save_before_associations", g.saveBeforeAssociationsCallback)

	g.ptrSize = len(fmt.Sprintf("%v", db))

	// Get Error converter
	g.converter, err = gormconv.New(db)
	if err != nil {
		return err
	}

	return nil
}

func (g *GORMRepository) log() unilogger.LeveledLogger {
	if g.logger == nil {
		logger := unilogger.NewBasicLogger(os.Stdout, "GORM Repository ", log.Ldate|log.Ltime|log.Lshortfile)
		logger.SetLevel(g.logLevel)
		g.logger = logger
	}
	return g.logger

}

func (g *GORMRepository) buildScopeGet(jsonScope *jsonapi.Scope) (*gorm.Scope, error) {
	gormScope := g.db.NewScope(jsonScope.Value)
	mStruct := gormScope.GetModelStruct()
	db := gormScope.DB()

	err := g.buildFilters(db, mStruct, jsonScope)
	if err != nil {
		return nil, err
	}

	// FieldSets
	if err = g.buildFieldSets(db, jsonScope, mStruct); err != nil {
		return nil, err
	}
	return gormScope, nil
}

func (g *GORMRepository) buildScopeList(jsonScope *jsonapi.Scope,
) (gormScope *gorm.Scope, err error) {
	gormScope = g.db.NewScope(jsonScope.GetValueAddress())
	db := gormScope.DB()

	mStruct := gormScope.GetModelStruct()

	// Filters
	err = g.buildFilters(db, mStruct, jsonScope)
	if err != nil {
		// fmt.Println(err.Error())
		return nil, err
	}

	// FieldSets
	if err = g.buildFieldSets(db, jsonScope, mStruct); err != nil {
		return
	}

	// Paginate
	buildPaginate(db, jsonScope)

	// Order
	if err = buildSorts(db, jsonScope, mStruct); err != nil {
		return
	}

	return gormScope, nil
}

// gets relationship from the database
func (g *GORMRepository) getRelationship(
	field *jsonapi.StructField,
	scope *jsonapi.Scope,
	gormScope *gorm.Scope,
) (err error) {
	var (
		fieldScope *gorm.Scope
		gormField  *gorm.StructField
		fkField    *gorm.Field

		errNilPrimary = errors.New("nil value")

		getBelongsToRelationship = func(singleValue, relationValue reflect.Value) error {
			// fmt.Printf("BelongsToRelationship for field: %v", fkField.Struct.Name)
			relationPrimary := relationValue.Elem().FieldByIndex(fieldScope.PrimaryField().Struct.Index)
			fkValue := singleValue.Elem().FieldByIndex(fkField.Struct.Index)

			if fkValue.Kind() == reflect.String {
				strFK := fkValue.Interface().(string)
				if strFK == "" {
					// fmt.Printf("Trying to get primary key for the belongs to relationship. It's empty. Field: %v\n", fkField.Struct.Name)
					return errNilPrimary
				}
			} else if !fkValue.IsValid() {
				// fmt.Printf("Field is not valid. %s", fkField.Struct.Name)
				return errNilPrimary
			}

			relationPrimary.Set(fkValue)
			return nil
		}

		// funcs
		getRelationshipSingle = func(singleValue reflect.Value) error {

			relationValue := singleValue.Elem().Field(field.GetFieldIndex())

			t := field.GetFieldType()
			switch t.Kind() {
			case reflect.Slice:
				sliceVal := reflect.New(reflect.SliceOf(t.Elem()))

				db := g.db.New()
				assoc := db.Model(singleValue.Interface()).
					Select(fieldScope.PrimaryField().DBName).
					Association(field.GetFieldName())

				if err := assoc.Error; err != nil {
					return err
				}

				relation := sliceVal.Interface()
				if err := assoc.Find(relation).Error; err != nil {
					return err
				}

				relationValue.Set(reflect.ValueOf(relation).Elem())

			case reflect.Ptr:

				relationValue.Set(reflect.New(t.Elem()))
				if fkField != nil {
					err = getBelongsToRelationship(singleValue, relationValue)
					if err != nil {
						return err
					}
					singleValue.Elem().Field(field.GetFieldIndex()).Set(relationValue)
					return nil
				} else {
					db := g.db.New()
					assoc := db.Model(singleValue.Interface()).
						Select(fieldScope.PrimaryField().DBName).
						Association(field.GetFieldName())

					if err := assoc.Error; err != nil {
						return err
					}

					relation := relationValue.Interface()
					err = assoc.Find(relation).Error

					if err != nil {
						if err == gorm.ErrRecordNotFound {
							relationValue.Set(reflect.Zero(t))
							return nil
						} else {
							return err
						}
					}
					relationValue.Set(reflect.ValueOf(relation))

				}
			}

			return nil
		}
	)

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
				err = fmt.Errorf("Unknown panic occured during getting scope's relationship.")
			}
		}
	}()

	fieldScope = g.db.NewScope(reflect.New(field.GetFieldType()).Elem().Interface())
	if fieldScope == nil {
		err := fmt.Errorf("Empty gorm scope for field: '%s' and model: '%v'.", field.GetFieldName(), scope.Struct.GetType())
		return err
	}

	// Get gormField as a gorm.StructField for given relationship field
	for _, gField := range gormScope.GetModelStruct().StructFields {
		if gField.Struct.Index[0] == field.GetFieldIndex() {
			gormField = gField
			break
		}
	}

	if gormField == nil {
		err := fmt.Errorf("No gormField for field: '%s'", field.GetFieldName())
		return err
	}

	// If given relationship is of Belongs_to type find a gorm
	if gormField.Relationship != nil && gormField.Relationship.Kind == annotationBelongsTo {
		for _, f := range gormScope.Fields() {
			if f.Name == gormField.Relationship.ForeignFieldNames[0] {
				fkField = f
			}
		}
		if fkField == nil {
			err := fmt.Errorf("No foreign field found for field: '%s'", gormField.Relationship.ForeignFieldNames[0])
			return err
		}
	}

	v := reflect.ValueOf(scope.Value)
	if v.Kind() == reflect.Slice {

		length := v.Len()
		for i := 0; i < length; i++ {

			singleValue := v.Index(i)
			// fmt.Printf("SingleValue: %v\n", singleValue.Type())
			err = getRelationshipSingle(singleValue)
			if err != nil {
				if err == errNilPrimary {
					if i == length-1 {
						v = v.Slice(0, i)
					} else {
						v = reflect.AppendSlice(v.Slice(0, i), v.Slice(i+1, length))
						i--
						length--
					}
				} else {
					return err
				}
			}

		}
	} else {
		// fmt.Printf("Get Single relationship: '%+v', '%v'", v.Kind())
		err = getRelationshipSingle(v)
		if err != nil {
			if err == errNilPrimary {
				return nil
			}
			return err
		}

	}

	return nil
}

func buildPaginate(db *gorm.DB, jsonScope *jsonapi.Scope) {
	if jsonScope.Pagination != nil {
		limit, offset := jsonScope.Pagination.GetLimitOffset()
		*db = *db.Limit(limit).Offset(offset)
	}
	return
}

// buildFieldSets helper for building FieldSets
func (g *GORMRepository) buildFieldSets(db *gorm.DB, jsonScope *jsonapi.Scope, mStruct *gorm.ModelStruct) error {

	var (
		fields    string
		foundPrim bool
	)

	for _, field := range jsonScope.Fieldset {
		if !field.IsRelationship() {
			index := field.GetFieldIndex()
			for _, gField := range mStruct.StructFields {
				if gField.Struct.Index[0] == index {
					if gField.IsIgnored {
						continue
					}

					if field.IsPrimary() {
						foundPrim = true
					}
					// this is the field
					fields += gField.DBName + ", "
				}
			}
		}
	}

	if !foundPrim {
		for _, primField := range mStruct.PrimaryFields {
			if isFieldEqual(primField, jsonScope.Struct.GetPrimaryField()) {
				fields = primField.DBName + ", " + fields
			}
		}
	}

	if len(fields) > 0 {
		fields = fields[:len(fields)-2]
	}
	*db = *db.Select(fields)
	return nil
}

func buildSorts(db *gorm.DB, jsonScope *jsonapi.Scope, mStruct *gorm.ModelStruct) error {

	for _, sort := range jsonScope.Sorts {
		if !sort.IsRelationship() {
			index := sort.GetFieldIndex()
			var sField *gorm.StructField
			if index == mStruct.PrimaryFields[0].Struct.Index[0] {
				sField = mStruct.PrimaryFields[0]
			} else {
				for _, gField := range mStruct.StructFields {
					if index == gField.Struct.Index[0] {
						sField = gField
					}
				}
			}
			if sField == nil {
				err := fmt.Errorf("Sort field: '%s' not found within model: '%s'", sort.GetFieldName(), mStruct.ModelType)

				return err
			}

			order := sField.DBName

			if sort.Order == jsonapi.DescendingOrder {
				order += " DESC"
			}
			*db = *db.Order(order)
		} else {
			// fmt.Println("Rel")
			// not implemented yet.
			// it should order the relationship id
			// and then make
		}
	}

	return nil
}

func (g *GORMRepository) getListRelationships(
	db *gorm.DB,
	scope *jsonapi.Scope,
) error {
	v := reflect.ValueOf(scope.Value)
	if v.Kind() != reflect.Slice {
		return errors.New("Provided value is not a slice")
	}

	if v.IsNil() {
		return errors.New("Nil value provided")
	}

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.IsNil() {
			continue
		}
		elemVal := elem.Interface()
		if err := g.getRelationships(db, scope, elemVal); err != nil {
			return err
		}

		v.Index(i).Set(reflect.ValueOf(elemVal))
	}
	return nil
}

func (g *GORMRepository) getRelationships(
	db *gorm.DB,
	scope *jsonapi.Scope,
	value interface{},
) error {

	gScope := db.New().NewScope(value)

	for _, field := range scope.Fieldset {
		if rel := field.GetRelationship(); rel != nil {
			switch rel.Kind {
			case jsonapi.RelBelongsTo:
				continue
			case jsonapi.RelHasMany, jsonapi.RelHasOne:
				if rel.Sync == nil || (rel.Sync != nil && *rel.Sync) {
					continue
				}
			case jsonapi.RelMany2Many:
				if rel.Sync != nil && *rel.Sync {
					continue
				}
			default:
				continue
			}
			tx := gScope.DB().Set("gorm:association:source", value)

			var gField *gorm.Field
			for _, gField = range gScope.Fields() {
				if isFieldEqual(gField.StructField, field) {
					break
				}
			}
			if gField == nil {
				g.log().Debug("Gorm field not found for: '%s'", field.GetFieldName())
				continue
			}

			if gField.IsIgnored {
				continue
			}

			if rel := gField.Relationship; rel != nil {
				switch rel.Kind {
				case "has_one":
					for idx, foreignKey := range rel.ForeignDBNames {
						if f, ok := gScope.FieldByName(rel.AssociationForeignDBNames[idx]); ok {
							tx = tx.Where(fmt.Sprintf("%v = ?", gScope.Quote(foreignKey)), f.Field.Interface())
						}
					}
					if rel.PolymorphicType != "" {
						tx = tx.Where(fmt.Sprintf("%v = ?", gScope.Quote(rel.PolymorphicDBName)), rel.PolymorphicValue)
					}

					if gField.Field.IsNil() {
						gField.Field.Set(reflect.New(gField.Field.Type().Elem()))
					}
					fValue := gField.Field.Interface()
					g.log().Debugf("Field Type: %v", gField.Field.Type().String())
					relScope := db.NewScope(fValue)

					// g.log().Debugf("HasOneQuery: %s", tx.Select(relScope.Quote(relScope.PrimaryKey())).SubQuery())
					g.log().Debugf("fValue: %+v", fValue)
					err := tx.Select(relScope.Quote(relScope.PrimaryKey())).Find(fValue).Error
					if err != nil {
						dbErr := g.converter.Convert(err)
						if !dbErr.Compare(unidb.ErrNoResult) {
							g.log().Errorf("Error while getting the relationship field: %s for model: %s. Err: %v", field.GetFieldName(), scope.Struct.GetType().Name(), err)
							return dbErr
						}
					}

					gField.Field.Set(reflect.ValueOf(fValue))
				case "has_many":
					for idx, foreignKey := range rel.ForeignDBNames {
						if f, ok := gScope.FieldByName(rel.AssociationForeignDBNames[idx]); ok {
							w := fmt.Sprintf("%v = ?", gScope.Quote(foreignKey))
							g.log().Debugf("Adding has_many where: '%s'", w)
							tx = tx.Where(w, f.Field.Interface())
						}
					}
					if rel.PolymorphicType != "" {
						tx = tx.Where(fmt.Sprintf("%v = ?", gScope.Quote(rel.PolymorphicDBName)), rel.PolymorphicValue)
					}

					fValue := gField.Field.Addr().Interface()
					relScope := db.NewScope(fValue)

					g.log().Debug("HasMany Query: %+v", tx.QueryExpr())
					err := tx.Select(relScope.PrimaryKey()).Find(fValue).Error
					if err != nil {
						dbErr := g.converter.Convert(err)
						if !dbErr.Compare(unidb.ErrNoResult) {
							g.log().Errorf("Error while getting the relationship field: %s for model: %s. Err: %v", field.GetFieldName(), scope.Struct.GetType().Name(), err)
							return dbErr
						}
					}

					gField.Field.Set(reflect.ValueOf(fValue).Elem())

				case "many_to_many":
					jth := rel.JoinTableHandler
					fValue := gField.Field.Addr().Interface()
					g.log().Debugf("Many2Many query: %v", jth.JoinWith(jth, tx, value).SubQuery())
					err := jth.JoinWith(jth, tx, value).Find(fValue).Error
					if err != nil {
						dbErr := g.converter.Convert(err)
						if !dbErr.Compare(unidb.ErrNoResult) {
							g.log().Errorf("Error while getting relation many2many: %s for model: %s. Err: %v", field.GetFieldName(), scope.Struct.GetType().Name(), err)
							return dbErr
						}
					}
					gField.Field.Set(reflect.ValueOf(fValue).Elem())
				default:
					continue

				}
			}

		}
	}
	return nil
}
