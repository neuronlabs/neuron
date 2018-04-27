package jsonapi

// IncludeScope is the includes information scope
// it contains the field to include from the root scope
// related subscope, and subfields to include.
type IncludeScope struct {
	// If field is included
	Field *StructField

	// RelatedScope is the scope where given include is described
	RelatedScope *Scope

	// if subField is included
	IncludedScopes []*IncludeScope
}
