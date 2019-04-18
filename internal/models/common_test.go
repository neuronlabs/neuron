package models

import (
	"github.com/neuronlabs/neuron/log"
	"testing"
)

type baseModel struct {
	ID         int    `neuron:"type=primary"`
	StringAttr string `neuron:"type=attr"`

	RelField   *embeddedModel `neuron:"type=relation;foreign=RelFieldID"`
	RelFieldID int            `neuron:"type=foreign"`
}

type embeddedModel struct {
	baseModel

	IntAttr int `neuron:"type=attr"`
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

func checkLogger() {
	if testing.Verbose() {
		log.SetLevel(log.LDEBUG)
	}
}
