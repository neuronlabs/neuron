package annotation

// Separators and other symbols.
const (
	// Separator is the symbol used to separate the sub-tags for given neuron tag.
	// Example: `neuron:"many2many=foreign,related_foreign"`
	//										 ^
	Separator = ","

	// TagSeparator is the symbol used to separate neuron based tags.
	// Example: `neuron:"type=attr;name=custom_name"`
	//								 ^
	TagSeparator = ";"
	// TagEqual is the symbol used to set the values for the for given neuron tag.
	// Example: `neuron:"type=attr"`
	//						    ^
	TagEqual = '='

	// NestedSeparator is the symbol used as a separator for the nested fields access.
	// Used in included or sort fields.
	// Example: field.relationship.
	// 				    ^
	NestedSeparator = "."

	// OpenedBracket is the symbol used in filtering system
	// which is used to open new logical part.
	// Example: filter[collection][name][$operator]
	//				  ^           ^     ^
	OpenedBracket = '['

	// ClosedBracket is the symbol used in filtering system
	// which is used to open new logical part.
	// Example: filter[collection][name][$operator]
	//				  			 ^     ^          ^
	ClosedBracket = ']'
)
