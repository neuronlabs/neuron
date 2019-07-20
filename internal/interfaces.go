package internal

// ForeignKeyModel is the interface used for getting and setting
// the value of foreign key fields for given model.
type ForeignKeyModel interface {
	ForeignKeyGetter
	ForeignKeySetter
}

// AttributeModel is the interface used for models that are allowed
// to get and set their attribute fields.
type AttributeModel interface {
	AttributeGetter
	AttributeSetter
}

// IDModel is the interface that is allowed to get and set the ID field.
type IDModel interface {
	IDGetter
	IDSetter
}

// IDGetter is the interface used to get the ID from the model instance.
type IDGetter interface {
	GetID() interface{}
}

// IDSetter is the interface used to set the ID for the model instance.
type IDSetter interface {
	SetID(id interface{}) error
}

// AttributeGetter is the interface used to get the attribute value
// from the model instance.
type AttributeGetter interface {
	GetAttribute(name string) interface{}
}

// AttributeSetter sets the attribute value for given model instance.
type AttributeSetter interface {
	SetAttribute(name string, value interface{}) error
}

// ForeignKeyGetter is the interface used to get the value of the
// foreign key field with given name from the model instance.
type ForeignKeyGetter interface {
	GetForeignKey(name string) interface{}
}

// ForeignKeySetter is the interface used to set the value of the foreign key field for the provided name
// for the model instance
type ForeignKeySetter interface {
	SetForeignKey(name string, value interface{}) error
}
