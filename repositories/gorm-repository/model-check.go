package gormrepo

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"reflect"
	"time"
)

func CheckGormModels(db *gorm.DB, models ...interface{}) error {
	return checkModels(db, models...)
}

func checkModels(db *gorm.DB, models ...interface{}) error {
	for _, model := range models {
		if err := checkSingleModel(model, db); err != nil {
			return err
		}
	}
	return nil
}

func checkSingleModel(model interface{}, db *gorm.DB) error {
	scope := db.NewScope(model)
	gormStruct := scope.GetModelStruct()
	for _, field := range gormStruct.StructFields {
		t := field.Struct.Type
		switch t.Kind() {
		case reflect.Ptr:
			t = t.Elem()
			if t.Kind() != reflect.Struct {
				continue
			}

		case reflect.Slice:
			t = t.Elem()
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if t.Kind() != reflect.Struct {
				continue
			}
		default:
			continue
		}

		if t == reflect.TypeOf(time.Time{}) {
			continue
		}

		if field.Relationship == nil {
			err := fmt.Errorf("Field: '%s' of type: '%v' does not have relatioship. Model: %v'",
				field.Name, field.Struct.Type, gormStruct.ModelType.Name())
			return err
		}

	}
	return nil

}
