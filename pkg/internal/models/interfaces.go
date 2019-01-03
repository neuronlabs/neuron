package models

type Collectioner interface {
	CollectionName() string
}

// SchemaNamer is the interface that defines the model's Schema Name
type SchemaNamer interface {
	SchemaName() string
}
