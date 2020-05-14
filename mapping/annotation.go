package mapping

// Model field's flag tags.
const (
	// AnnotationFlags is the neuron model field's tag used for defining field flags.
	AnnotationFlags = "flags"
	// AnnotationHidden defines that the field should be hidden from marshaling.
	AnnotationHidden = "hidden"
	// AnnotationISO8601 sets the time field format to ISO8601.
	AnnotationISO8601 = "iso8601"
	// AnnotationOmitEmpty allows to omit marshaling this field if it's zero-value.
	AnnotationOmitEmpty = "omitempty"
	// AnnotationI18n defines that this field is internationalization ready.
	AnnotationI18n = "i18n"
	// AnnotationNoFilter is the neuron model field's flag that disallows to query filter for given field.
	AnnotationNoFilter = "nofilter"
	// AnnotationNotSortable is the neuron model field's flag that disallows to query sort on given field.
	AnnotationNotSortable = "nosort"
	// AnnotationCreatedAt is the neuron model field's flag that defines CreatedAt field.
	AnnotationCreatedAt = "created_at"
	// AnnotationDeletedAt is the neuron model field's flag that defines DeletedAt field.
	AnnotationDeletedAt = "deleted_at"
	// AnnotationUpdatedAt is the neuron model field's flag that defines UpdatedAt field.
	AnnotationUpdatedAt = "updated_at"
)

// AnnotationNeuron is the root struct field annotation tag.
const AnnotationNeuron = "neuron"

// Model primary field annotation tags.
const (
	AnnotationPrimary      = "primary"
	AnnotationPrimaryFull  = "primary_key"
	AnnotationPrimaryFullS = "primarykey"
	AnnotationID           = "id"
	AnnotationPrimaryShort = "pk"
)

// AnnotationClientID states if the primary field could be defined by the client.
const AnnotationClientID = "client-id"

// Model attribute field annotation tags.
const (
	AnnotationAttribute     = "attr"
	AnnotationAttributeFull = "attribute"
)

// AnnotationLanguage defines the attribute field that contains the language tag.
// for i18n.
const AnnotationLanguage = "lang"

// Model relationship field annotation tags.
const (
	AnnotationRelation     = "relation"
	AnnotationRelationFull = "relationship"
)

const (
	// AnnotationName is the neuron model field's tag used to set the NeuronName.
	AnnotationName = "name"
	// AnnotationFieldType is the neuron model field's tag used to set the neuron field type.
	AnnotationFieldType = "type"
	// AnnotationNestedField is the model field's neuron tag that defines if the field type is of nested type.
	AnnotationNestedField = "nested"
)

// AnnotationManyToMany is the neuron relationship field tag that states this relationship is of type many2many.
const AnnotationManyToMany = "many2many"

// Model foreign key field annotation tags.
const (
	AnnotationForeignKey      = "foreign"
	AnnotationForeignKeyFull  = "foreign_key"
	AnnotationForeignKeyFullS = "foreignkey"
	AnnotationForeignKeyShort = "fk"
)

// Separators and other symbols.
const (
	// AnnotationSeparator is the symbol used to separate the sub-tags for given neuron tag.
	// Example: `neuron:"many2many=foreign,related_foreign"`
	//										 ^
	AnnotationSeparator = ","

	// AnnotationTagSeparator is the symbol used to separate neuron based tags.
	// Example: `neuron:"type=attr;name=custom_name"`
	//								 ^
	AnnotationTagSeparator = ";"

	// AnnotationTagEqual is the symbol used to set the values for the for given neuron tag.
	// Example: `neuron:"type=attr"`
	//						    ^
	AnnotationTagEqual = '='

	// AnnotationNestedSeparator is the symbol used as a separator for the nested fields access.
	// Used in included or sort fields.
	// Example: field.relationship.
	// 				    ^
	AnnotationNestedSeparator = "."

	// AnnotationOpenedBracket is the symbol used in filtering system
	// which is used to open new logical part.
	// Example: filter[collection][name][$operator]
	//				  ^           ^     ^
	AnnotationOpenedBracket = '['

	// AnnotationClosedBracket is the symbol used in filtering system
	// which is used to open new logical part.
	// Example: filter[collection][name][$operator]
	//				  			 ^     ^          ^
	AnnotationClosedBracket = ']'
)
