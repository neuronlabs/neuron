package input

import (
	"sort"

	"github.com/neuronlabs/inflection"
	"github.com/neuronlabs/strcase"
)

// MultiModel is a input for the multi model templates.
type MultiModel struct {
	PackageName string
	Imports     Imports
	Models      []*Model
}

// Model is a structure used to insert model into template.
type Model struct {
	TestFile                                   bool
	FileName                                   string
	PackageName                                string
	Imports                                    Imports
	Name                                       string
	Receiver                                   string
	CollectionName                             string
	RepositoryName                             string
	Primary                                    *Field
	Fields                                     []*Field
	Fielder, SingleRelationer, MultiRelationer bool
	Relations                                  []*Field
	Receivers                                  map[string]int
}

// AddImport adds 'imp' imported package if it doesn't exists already.
func (m *Model) AddImport(imp string) {
	m.Imports.Add(imp)
}

// Collection returns model's collection.
func (m *Model) Collection() *Collection {
	return &Collection{
		Name:         strcase.ToLowerCamel(inflection.Plural(m.Name)),
		VariableName: strcase.ToCamel(inflection.Plural(m.Name)),
		QueryBuilder: strcase.ToLowerCamel(inflection.Plural(m.Name) + "QueryBuilder"),
	}
}

// CollectionInput returns template collection input for given model.
func (m *Model) CollectionInput(packageName string) *CollectionInput {
	c := &CollectionInput{
		PackageName: packageName,
		Imports: []string{
			"context",
			"github.com/neuronlabs/neuron/controller",
			"github.com/neuronlabs/neuron/errors",
			"github.com/neuronlabs/neuron/mapping",
			"github.com/neuronlabs/neuron/query",
		},
		Model:      m,
		Collection: m.Collection(),
	}
	if c.PackageName == "" {
		c.PackageName = m.PackageName
	}
	return c
}

// SortFields sorts the fields in the model.
func (m *Model) SortFields() {
	sort.Slice(m.Fields, func(i, j int) bool {
		return m.Fields[i].Index < m.Fields[j].Index
	})
}
