package reflectgen

import (
	"github.com/neuronlabs/neuron/mapping"
)

// ModelTemplateInput is the input for the model templates.
type ModelTemplateInput struct {
	PackageName string
	Collection  Collection
	Model       Model
	Imports     []string
}

func (m *ModelTemplateInput) addImport(imp string) {
	var found bool
	for _, imp2 := range m.Imports {
		if imp2 == imp {
			found = true
			break
		}
	}
	if !found {
		m.Imports = append(m.Imports, imp)
	}
}

// NeuronCollectionName is a structure used to insert into collection definition.
type Collection struct {
	Name         string
	VariableName string
	Receiver     string
	QueryBuilder string
}

// Model is a structure used to insert model into template.
type Model struct {
	Name                                              string
	Receiver                                          string
	Collection                                        string
	RepositoryName                                    string
	Primary                                           *Field
	Fields                                            []*Field
	Fielder, SingleRelationer, MultiRelationer        bool
	Attributes, ForeignKeys, Relations, PrivateFields []*Field
}

// Field is a structure used to insert into model field template.
type Field struct {
	Index                          int
	Name                           string
	Type                           string
	BeforeZero                     string
	AfterZero                      string
	Zero                           string
	AlternateTypes                 []string
	Scanner, Sortable, ZeroChecker bool
	Tags                           string

	Model *Model

	NeuronField *mapping.StructField
}

// IsZero returns string template IsZero checker.
func (f *Field) IsZero() string {
	if f.ZeroChecker {
		return f.Model.Receiver + f.Name + ".IsZero()"
	}
	return f.BeforeZero + f.Model.Receiver + f.Name + f.AfterZero
}

// GetZero gets the zero value string for given field.
func (f *Field) GetZero() string {
	if f.ZeroChecker {
		return f.Model.Receiver + f.Name + ".GetZero()"
	}
	return f.Zero
}
