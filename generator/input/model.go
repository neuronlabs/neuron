package input

import (
	"sort"

	"github.com/neuronlabs/inflection"
	"github.com/neuronlabs/strcase"
)

// Model is a structure used to insert model into template.
type Model struct {
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

// NeuronCollectionName returns model's collection.
func (m *Model) Collection() *Collection {
	return &Collection{
		Name:         strcase.ToLowerCamel(inflection.Plural(m.Name)),
		VariableName: strcase.ToCamel(inflection.Plural(m.Name)),
		QueryBuilder: strcase.ToLowerCamel(inflection.Plural(m.Name) + "QueryBuilder"),
	}
}

// SortFields sorts the fields in the model.
func (m *Model) SortFields() {
	sort.Slice(m.Fields, func(i, j int) bool {
		return m.Fields[i].Index < m.Fields[j].Index
	})
}
