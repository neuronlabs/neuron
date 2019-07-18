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
	AnnotationPrimaryShort = "pk"

	AnnotationClientID = "client-id"
	AnnotationLanguage = "lang"

	// attributes
	AnnotationAttribute     = "attr"
	AnnotationAttributeFull = "attribute"

	AnnotationRelation     = "relation"
	AnnotationRelationFull = "relationship"

	AnnotationFilterKey = "filterkey"

	AnnotationForeignKey      = "foreign"
	AnnotationForeignKeyFull  = "foreign_key"
	AnnotationForeignKeyFullS = "foreignkey"
	AnnotationForeignKeyShort = "fk"

	AnnotationNestedField = "nested"

	// relation tag
	AnnotationManyToMany = "many2many"

	AnnotationOnCreate = "on_create"
	AnnotationOnPatch  = "on_patch"
	AnnotationOnDelete = "on_delete"
	AnnotationOrder    = "order"
	AnnotationOnError  = "on_error"
	AnnotationOnChange = "on_change"

	AnnotationRelationRestrict = "restrict"
	AnnotationRelationNoAction = "no-action"
	AnnotationRelationCascade  = "cascade"
	AnnotationRelationSetNull  = "set-null"

	AnnotationFailOnError     = "fail"
	AnnotationContinueOnError = "continue"

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

	IsPointerTime = "jsonapi:is-ptr-time"
)
