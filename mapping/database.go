package mapping

import (
	"fmt"
	"strings"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
)

// DatabaseIndex is structure that defines database index with it's name, mapped fields, uniqueness and type.
type DatabaseIndex struct {
	// Name is the name of the index.
	Name string
	// Type is the database index type.
	Type string
	// Parameters are the index defined parameters.
	Parameters []string
	// Unique defines if the index is unique.
	Unique bool
	// Fields defined fields mapped for this index.
	Fields []*StructField
}

// By default the model's database equivalent name (table - in SQL) is the model's collection name.
// For custom database model's name it needs to implement DatabaseNamer interface.
//
// Some database implementations allow to have different schema in a single database.
// If a model is stored in schema other than default, than it needs to implement DatabaseSchemaNamer interface.
//
// Each field could be customized in a database by specifying struct field tags.
// Database struct field tag is defined with'db':
// - It's subtags are split with ';'.
// - The first subtag is always database name - if the name is empty or '_' then it's value would not be set.
// - The other subtags could be composed in two ways:
// 		subtag_name:subtag_value1,subtag_value2 OR subtag_name.
// - Subtag values are split using comma ','.
//
// Neuron defined database subtags:
//	'-' 			- 	field is skipped in the database.
//	'unique'		- 	field has a unique constraint.
// 	'index'			- 	defines field's index. If no values defined than it would take default index name for given field.
//						The first subvalue always define index name.
//						All other properties would get stored in the 'Parameters'.
// 	'unique_index'	- 	defines field's unique index if no values defined than it would take default index name for given field.
//						The first subvalue always define index name.
//						All other properties would get stored in the 'Parameters'.
// 	'type' 			-	defines database field equivalent 'type' this depends on given database implementation.
//	'notnull'		- 	this field is marked as 'not null' in the database.
//	'null'			- 	this field is marked as 'null'  in the database, used in case when the defaultNotNull flag is set true.
//
//
// Example database tag:
// db:"_;unique_index=,btree" 		- no database name defined, defined unique index with the parameter 'btree'
// db:";unique_index" 		- no database name defined, defined unique index with the parameter 'btree'
// db:"db_name;index=index_name" 	- database name defined, index with given name defined.
// db:"_;index"						- no database name defined, default index name for given field.
func (m *ModelStruct) extractDatabaseTags(defaultNotNull bool) error {
	model := NewModel(m)
	if namer, ok := model.(DatabaseSchemaNamer); ok {
		m.DatabaseSchemaName = namer.DatabaseSchemaName()
	}
	if namer, ok := model.(DatabaseNamer); ok {
		m.DatabaseName = namer.DatabaseName()
	}

	namedIndexes := map[string]*DatabaseIndex{}
	for _, field := range m.fields {
		// If the defaultNotNull is set and the field is not a pointer set it to notnull.
		if defaultNotNull && !field.IsPtr() {
			field.fieldFlags |= fDatabaseNotNull
		}
		tags := field.extractFieldTags("db", AnnotationTagSeparator, AnnotationSeparator)
		if len(tags) == 0 {
			continue
		}

		// The first tag is always database name.
		name := tags[0]
		if name.Key == "-" {
			field.fieldFlags |= fDatabaseSkip
			continue
		}
		switch name.Key {
		case "name":
			if len(name.Values) != 1 {
				return errors.Wrapf(ErrMapping, "model's: %s field: '%s' database tag name defined without value", m, field)
			}
			field.DatabaseName = name.Values[0]
		case "unique_index", "index", "type", "notnull", "null", "unique":
			tags = append(tags, name)
		case "_":
		default:
			field.DatabaseName = name.Key
		}

		for _, tag := range tags[1:] {
			switch tag.Key {
			case "index", "unique_index":
				index := &DatabaseIndex{
					Unique: tag.Key == "unique_index",
				}
				var defaultName bool
				if len(tag.Values) == 0 || (len(tag.Values) > 0 && tag.Values[0] == "") {
					defaultName = true
					index.Name = fmt.Sprintf("idx_nrn_%s_%s", m.collection, field.DatabaseName)
					if index.Unique {
						index.Name += "_unique"
					}
					index.Fields = []*StructField{field}
				}
				if len(tag.Values) > 0 {
					if !defaultName {
						index.Name = tag.Values[0]
						if _, ok := namedIndexes[index.Name]; ok {
							index = namedIndexes[index.Name]
							index.Fields = append(index.Fields, field)
						} else {
							namedIndexes[index.Name] = index
							index.Fields = []*StructField{field}
						}
					}
					if len(tag.Values) > 0 {
						for _, value := range tag.Values[1:] {
							for _, param := range index.Parameters {
								if param == value {
									continue
								}
								index.Parameters = append(index.Parameters, param)
							}
						}
					}
				}
			case "type":
				switch len(tag.Values) {
				case 0:
					return errors.Wrapf(ErrMapping, "database field: '%s' type value not defined.", field.Name())
				case 1:
				default:
					tag.Values[0] = strings.Join(tag.Values, ",")
				}
				field.DatabaseType = tag.Values[0]
			case "notnull":
				field.fieldFlags |= fDatabaseNotNull
				if len(tag.Values) > 0 {
					return errors.Wrap(ErrMapping, "database notnull constraint can't use values")
				}
			case "null":
				if field.fieldFlags&fDatabaseNotNull != 0 {
					field.fieldFlags &= ^fDatabaseNotNull
				}
				if len(tag.Values) > 0 {
					return errors.Wrap(ErrMapping, "database null constraint can't use values")
				}
			case "unique":
				field.fieldFlags |= fDatabaseUnique
				if len(tag.Values) > 0 {
					return errors.Wrap(ErrMapping, "database unique constraint can't use values")
				}
			default:
				field.DatabaseUnknownTags = append(field.DatabaseUnknownTags, tag)
				log.Debug("Unknown database tag: %s", tag.Key)
			}
		}
	}

	for _, index := range namedIndexes {
		m.databaseIndexes = append(m.databaseIndexes, index)
	}
	return nil
}
