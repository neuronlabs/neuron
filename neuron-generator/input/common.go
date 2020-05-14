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
	ExternalController bool
}

// CollectionInput creates a collection input.
type CollectionInput struct {
	PackageName        string
	Imports            Imports
	Model              *Model
	Collection         *Collection
	ExternalController bool
}

// MultiCollectionInput is the input for multiple collections.
type MultiCollectionInput struct {
	PackageName        string
	Imports            Imports
	Collections        []*CollectionInput
	ExternalController bool
}

// Collection is a structure used to insert into collection definition.
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
