package input

// Imports is the wrapper over the string slice that allows to add exclusive and sort given import packages.
type Imports []string

// Add if the 'pkg' doesn't exists in the imports the function inserts the package into the slice.
func (i *Imports) Add(pkg string) {
	var found bool
	for _, imp2 := range *i {
		if imp2 == pkg {
			found = true
			break
		}
	}
	if !found {
		*i = append(*i, pkg)
	}
}

// Sort sorts related imports.
func (i *Imports) Sort() {
	// TODO: apply goimports sort.
}

// Collections is template input structure used for creation of the collections initialization file.
type Collections struct {
	PackageName        string
	Imports            Imports
	ExportedController bool
}

// CollectionInput creates a collection input.
type CollectionInput struct {
	PackageName        string
	Imports            Imports
	Model              *Model
	Collection         *Collection
	ExternalController bool
}

// CollectionInput returns template collection input for given model.
func (m *Model) CollectionInput(packageName string) *CollectionInput {
	c := &CollectionInput{
		PackageName: packageName,
		Model:       m,
		Imports: []string{
			"github.com/neuronlabs/neuron/controller",
			"github.com/neuronlabs/neuron/query",
			"github.com/neuronlabs/neuron/mapping",
		},
	}
	for _, field := range m.Fields {
		if field.Selector != "" {

		}
	}
	return c
}

// NeuronCollectionName is a structure used to insert into collection definition.
type Collection struct {
	// Name is the lowerCamelCase plural name of the model.
	Name string
	// VariableName is the CamelCase plural name of the model.
	VariableName string
	// QueryBuilder is the name of the query builder for given collection.
	QueryBuilder string
}

// Receiver gets the first letter from the collection name - used as the function receiver.
func (c Collection) Receiver() string {
	return c.Name[:1]
}

// ZeroChecker is the interface that allows to check if the value is zero.
type ZeroChecker interface {
	IsZero() bool
	GetZero() interface{}
}