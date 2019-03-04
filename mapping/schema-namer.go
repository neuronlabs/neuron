package mapping

// SchemaNamer is the interface used for mapping the models within the controller.
type SchemaNamer interface {
	SchemaName() string
}
