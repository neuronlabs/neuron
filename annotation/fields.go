package annotation

// Neuron is the root structfield annotation tag.
const Neuron = "neuron"

// Model primary field annotation tags.
const (
	Primary      = "primary"
	PrimaryFull  = "primary_key"
	PrimaryFullS = "primarykey"
	ID           = "id"
	PrimaryShort = "pk"
)

// ClientID states if the primary field could be defined by the client.
const ClientID = "client-id"

// Model attribute field annotation tags.
const (
	Attribute     = "attr"
	AttributeFull = "attribute"
)

// Language defines the attribute field that contains the language tag.
// for i18n.
const Language = "lang"

// Model relationship field annotation tags.
const (
	Relation     = "relation"
	RelationFull = "relationship"
)

const (
	// Name is the neuron model field's tag used to set the NeuronName.
	Name = "name"
	// FieldType is the neuron model field's tag used to set the neuron field type.
	FieldType = "type"
)

// ManyToMany is the neuron relationship field tag that states this relationship is of type many2many.
const ManyToMany = "many2many"

// Model foreign key field annotation tags.
const (
	ForeignKey      = "foreign"
	ForeignKeyFull  = "foreign_key"
	ForeignKeyFullS = "foreignkey"
	ForeignKeyShort = "fk"
)

const (
	// FilterKey is the model's filter key field tag.
	FilterKey = "filterkey"
	// NestedField is the model field's neuron tag that defines if the field type is of nested type.
	NestedField = "nested"
)
