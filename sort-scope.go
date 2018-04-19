package jsonapi

type Order int

const (
	AscendingOrder Order = iota
	DescendingOrder
)

// SortScope is a field that describes the sorting rules for given
type SortScope struct {
	Field *StructField
	// Order defines if the sorting order (ascending or descending)
	Order Order
	// RelScopes is the relationship sort scopes
	RelScopes []*SortScope
}

func (s *SortScope) setRelationScope(sortSplitted []string, order Order) (invalidField bool) {
	var (
		relScope *SortScope
		sField   *StructField
	)
	if s.Field.jsonAPIType != RelationshipSingle && s.Field.jsonAPIType != RelationshipMultiple {
		invalidField = true
		return
	}
	// check attribute
	switch len(sortSplitted) {
	case 0:
		invalidField = true
		return
	case 1:
		sort := sortSplitted[0]
		if sort == annotationID {
			sField = s.Field.relatedStruct.primary
		} else {
			sField = s.Field.relatedStruct.attributes[sortSplitted[0]]
			if sField == nil {
				invalidField = true
				return
			}
		}

		s.RelScopes = append(s.RelScopes, &SortScope{Field: sField, Order: order})
	default:
		sField := s.Field.relatedStruct.relationships[sortSplitted[0]]
		if sField == nil {
			invalidField = true
			return
		}

		for i := range s.RelScopes {
			if s.RelScopes[i].Field.getFieldIndex() == sField.getFieldIndex() {
				relScope = s.RelScopes[i]
				break
			}
		}
		if relScope == nil {
			relScope = &SortScope{Field: sField}
		}
		invalidField = relScope.setRelationScope(sortSplitted[1:], order)
		if !invalidField {
			s.RelScopes = append(s.RelScopes, relScope)
		}
		return
	}

	return
}
