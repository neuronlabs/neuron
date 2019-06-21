package internal

// constants defined for the internal neuron packages
const (
	// StructTag Annotation strings
	AnnotationNeuron = "neuron"

	// type tag
	AnnotationPrimary      = "primary"
	AnnotationPrimaryFull  = "primary_key"
	AnnotationPrimaryFullS = "primarykey"
	AnnotationID           = "id"

	AnnotationClientID = "client-id"
	AnnotationLanguage = "langtag"

	// attributes
	AnnotationAttribute     = "attr"
	AnnotationAttributeFull = "attribute"

	AnnotationRelation     = "relation"
	AnnotationRelationFull = "relationship"

	AnnotationFilterKey = "filterkey"

	AnnotationForeignKey      = "foreign"
	AnnotationForeignKeyFull  = "foreign_key"
	AnnotationForeignKeyFullS = "foreignkey"

	AnnotationNestedField = "nested"

	// relation tag
	AnnotationManyToMany = "many2many"

	AnnotationRelationRestrict = "restrict"
	AnnotationRelationNoAction = "no-action"
	AnnotationRelationCascade  = "cascade"
	AnnotationRelationSetNull  = "set-null"

	AnnotationDefault = "default"

	// name tag
	AnnotationName = "name"

	// flags tag
	AnnotationHidden      = "hidden"
	AnnotationISO8601     = "iso8601"
	AnnotationOmitEmpty   = "omitempty"
	AnnotationI18n        = "i18n"
	AnnotationFieldType   = "type"
	AnnotationFlags       = "flags"
	AnnotationNoFilter    = "nofilter"
	AnnotationNotSortable = "nosort"

	AnnotationSeperator         = ","
	AnnotationRelationSeperator = ":"
	AnnotationTagSeperator      = ";"
	AnnotationTagEqual          = '='

	AnnotationNestedSeperator = "."
	AnnotationOpenedBracket   = '['
	AnnotationClosedBracket   = ']'

	// FILTERS
	// disable for filtering purpose

	// not used currently
	// AnnotationNot = "not"
	// AnnotationOr  = "or"
	// AnnotationAnd = "and"

	// string only filters

	Iso8601TimeFormat = "2006-01-02T15:04:05Z"

	IsPointerTime = "jsonapi:is-ptr-time"
)
