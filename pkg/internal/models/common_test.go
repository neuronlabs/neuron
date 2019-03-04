package models

type baseModel struct {
	ID         int    `jsonapi:"type=primary"`
	StringAttr string `jsonapi:"type=attr"`

	RelField   *embeddedModel `jsonapi:"type=relation;foreign=RelFieldID"`
	RelFieldID int            `jsonapi:"type=foreign"`
}

type embeddedModel struct {
	baseModel

	IntAttr int `jsonapi:"type=attr"`
}

// RepositoryName returns the name of the repository for the embeddedModel.
// Implements RepositoryNamer interface.
func (m *embeddedModel) RepositoryName() string {
	return defaultRepo
}

// SchemaName returns the name of the schema for the embeddedModel.
// Implements SchemaNamer interface.
func (m *embeddedModel) SchemaName() string {
	return defaultSchema
}
