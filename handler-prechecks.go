package jsonapi

/**

GET/LIST Precheck Filters

	Precheck filters should check the values that are provided in the filter fields
	or that are used after

*/

func (h *Handler) QueryPrecheckPair(
	scope *Scope,
	pairs ...*PresetPair,
) error {

	// iterate over precheck pairs
	// if the precheck filter matches any filter on the provided scope
	// check the values that the
	for _, pair := range pairs {
		precheckScope, precheckFilter := pair.GetPair()
		var found bool
		switch precheckFilter.GetFieldKind() {
		case Primary:
			for _, primFilter := range scope.PrimaryFilters {
				if primFilter.GetFieldIndex() == precheckFilter.GetFieldIndex() {
					found = true
					break
				}
			}
		case Attribute:
			for _, attrFilter := range scope.AttributeFilters {
				if attrFilter.GetFieldIndex() == precheckFilter.GetFieldIndex() {
					found = true
					break
				}
			}
		case RelationshipMultiple, RelationshipSingle:
			for _, relationFilter := range scope.RelationshipFilters {
				if relationFilter.GetFieldIndex() == precheckFilter.GetFieldIndex() {
					found = true
					break
				}
			}
		}
		if !found {
			continue
		}
		// do the business
		h.log.Debugf("Have to use scope: %v", precheckScope)

	}
	return nil
}
