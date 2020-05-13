package input

import "strings"

// Field is a structure used to insert into model field template.
type Field struct {
	Index            int
	Name, NeuronName string
	// Type is current field Type for given model.
	Type                                           string
	BeforeZero, AfterZero, Zero                    string
	AlternateTypes, WrappedTypes                   []string
	Scanner, Sortable, ZeroChecker                 bool
	Tags                                           string
	IsPointer, IsElemPointer, IsSlice, IsByteSlice bool
	// Selector is the import package name for given field
	// i.e.:
	// 	for field - time.Time - the Type would be 'Time' and Selector 'time'.
	Selector string

	Model *Model
}

// BaseType returns field type without pointer or slices.
func (f *Field) BaseType() string {
	index := strings.LastIndexAny(f.Type, "[]*")
	if index == -1 {
		return f.Type
	}
	return f.Type[index+1:]
}

// IsZero returns string template IsZero checker.
func (f *Field) IsZero() string {
	if f.ZeroChecker {
		return f.Model.Receiver + "." + f.Name + ".IsZero()"
	}
	return f.BeforeZero + f.Model.Receiver + "." + f.Name + f.AfterZero
}

// GetZero gets the zero value string for given field.
func (f *Field) GetZero() string {
	if f.ZeroChecker {
		return f.Model.Receiver + f.Name + ".GetZero()"
	}
	return f.Zero
}
