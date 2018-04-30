package jsonapi

// IncludeScope is the includes information scope
// it contains the field to include from the root scope
// related subscope, and subfields to include.
type IncludeField struct {
	*StructField

	// RelatedScope is the scope where given include is described
	Scope *Scope

	// if subField is included
	IncludedSubfields []*IncludeField
}

func newIncludeField(field *StructField, scope *Scope) *IncludeField {
	includeField := new(IncludeField)
	includeField.StructField = field
	includeField.Scope = scope
	return includeField
}

// check for relationship in included name
// if nested, then separate it.
// all nested fields are added into included field subfields
// the scope built for the included field is added into IncludedScopes so as the subfield scopes.
// the IncludedField is added to the container IncludedFields
// the scope does not change for the subfields
// func newIncludedField(included string, model *ModelStruct, scope *Scope,
// ) (includeField *IncludeField, errs []*ErrorObject) {
// 	var relationScope *Scope

// 	relatedField, ok := model.relationships[included]
// 	if !ok {
// 		// no relationship found check nesteds
// 		index := strings.Index(included, annotationNestedSeperator)
// 		if index == -1 {
// 			errs = append(errs, errNoRelationship(model.collectionType, included))
// 			return
// 		}

// 		// root part of included (root.subfield)
// 		root := included[:index]
// 		relatedField, ok := model.relationships[root]
// 		if !ok {
// 			errs = append(errs, errNoRelationship(model.collectionType, root))
// 			return
// 		}

// 		// Check if no other scope for the related collection exists
// 		relatedModel := relatedField.relatedStruct
// 		relationScope = scope.IncludedScopes[relatedMStruct]
// 		if relationScope == nil {
// 			var isSlice bool
// 			if relatedField.jsonAPIType == RelationshipMultiple {
// 				isSlice = true
// 			}
// 			relationScope = newScope(relatedModel)
// 			scope.IncludedScopes[relatedModel] = sub
// 		}
// 		// subfield is the nested include inside root field
// 		subfield := included[index+1:]
// 		relatedSubField, ok := relatedModel.relationships[subfield]
// 		if !ok {
// 			// if no subfield found error should occur
// 			errs = append(errs, errNoRelationship(relatedModel.collectionType, subfield))
// 			return
// 		}

// 		// subfield is found so we can check if root included was already included

// 		// find include scope if already used

// 		for _, inc := range scope.IncludedFields {
// 			if inc.getFieldIndex() == relatedField.getFieldIndex() {
// 				includeField = inc
// 			}
// 		}

// 		// if no includedField for root found
// 		// create new included Field
// 		if includeField == nil {
// 			includeField = &IncludeField{StructField: relatedField, Scope: relationScope}
// 			scope.Included = append(scope.Included, includeField)
// 		}

// 		var nestedInclude *IncludeField
// 		nestedInclude, errs = newIncludedField(included[index+1:], scope, relatedModel)

// 		var isInFieldSet bool
// 		for _, field := range relationScope.Fieldset {
// 			if field.getFieldIndex() == nestedInclude.getFieldIndex() {
// 				isInFieldSet = true
// 			}
// 		}

// 		if !isInFieldSet {
// 			relationScope.Fieldset = append(relationScope.Fieldset, nestedInclude)
// 		}
// 	} else {
// 		relationScope, ok = scope.IncludedScopes[relatedField.relatedStruct]
// 		if !ok {
// 			relationScope = newScope(relatedField.relatedStruct)
// 			scope.IncludedScopes[relatedField.relatedStruct] = relationScope
// 		}
// 	}

// }

func (i *IncludeField) buildNestedInclude(nested string, scope *Scope,
) (errs []*ErrorObject) {

	var (
		nestedInclude *IncludeField

		relationField *StructField
		ok            bool
	)

	// check in the field's model if it contains this nested include
	relationField, ok = i.mStruct.relationships[nested]
	if !ok {
		// if no relationship found, then check if it is possible to separate dots from 'nested'
		// no relationship found check nesteds
		index := strings.Index(included, annotationNestedSeperator)
		if index == -1 {
			errs = append(errs, errNoRelationship(i.mStruct.collectionType, included))
			return
		}

		// field part of included (field.subfield)
		field := included[:index]
		relationField, ok = i.mStruct.relationships[field]
		if !ok {
			// still not found - then add error and return
			errs = append(errs, errNoRelationship(i.mStruct.collectionType, field))
			return
		}

		nestedInclude = i.getOrCreateNestedInclude(relationField)
		// build recursively nested fields
		errs = append(errs, nestedInclude.buildNestedInclude(nested[index+1:], scope)...)
	} else {
		nestedInclude = i.getOrCreateNestedInclude(relationField)
	}

	_ = scope.createIncludedScope(nestedInclude.relatedStruct)
	return
}

// getOrCreateNestedInclude - get from includedSubfiedls or if no such field
// create new included.
func (i *IncludeField) getOrCreateNestedInclude(field *StructField) *IncludeField {
	for _, subfield := range i.IncludedSubfields {
		if subfield.getFieldIndex() == field.getFieldIndex() {
			return subfield
		}
	}
	includeField := new(IncludeField)
	includeField.StructField = field
	i.IncludedSubfields = append(i.IncludedSubfields, includeField)
	return includeField
}

func (i *IncludeField) includeSubfield(includeField *IncludeField) {
	i.IncludedSubfields = append(i.IncludedSubfields, includeField)
}
