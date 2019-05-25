package scope

import (
	"context"
	"fmt"
	"github.com/kucjac/uni-db"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"sync"

	"github.com/neuronlabs/neuron/internal/query/filters"
	scope "github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"reflect"
)

var (
	// ProcessGetIncluded is the process that gets the included scope values
	ProcessGetIncluded = &Process{
		Name: "neuron:get_included",
		Func: getIncludedFunc,
	}

	// ProcessSetBelongsToRelationships is the Process that sets the BelongsToRelationships
	ProcessSetBelongsToRelationships = &Process{
		Name: "neuron:set_belongs_to_relationships",
		Func: setBelongsToRelationshipsFunc,
	}

	// ProcessGetForeignRelationships is the Process that gets the foreign relationships
	ProcessGetForeignRelationships = &Process{
		Name: "neuron:get_foreign_relationships",
		Func: getForeignRelationshipsFunc,
	}

	// ProcessConvertRelationshipFilters converts the relationship filters into a primary or foreign key filters of the root scope
	ProcessConvertRelationshipFilters = &Process{
		Name: "neuron:convert_relationship_filters",
		Func: convertRelationshipFilters,
	}
)

func convertRelationshipFilters(ctx context.Context, s *Scope) error {
	log.Debugf("SCOPE[%s] convertRelationshipFilters", s.ID())
	relationshipFilters := (*scope.Scope)(s).RelationshipFilters()

	if len(relationshipFilters) == 0 {
		return nil
	}

	wg := &sync.WaitGroup{}

	bufSize := 10

	if len(relationshipFilters) < bufSize {
		bufSize = len(relationshipFilters)
	}

	var results = make(chan interface{}, bufSize)

	// create the cancelable context for the sub context
	maxTimeout := s.Controller().Config.Builder.RepositoryTimeout
	for _, rel := range relationshipFilters {
		if rel.StructField().Relationship().Struct().Config() == nil {
			continue
		}
		if modelRepo := rel.StructField().Relationship().Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	for i, rf := range (*scope.Scope)(s).RelationshipFilters() {

		// cast the filter to the internal structure definition
		filter := (*filters.FilterField)(rf)

		// check the relatinoship kind
		switch rf.StructField().Relationship().Kind() {
		case models.RelBelongsTo:
			go handleBelongsToRelationshipFilter(ctx, s, i, filter, results)

		case models.RelHasMany:
			// these shoud contain the foreign key in their definitions
			go handleHasManyRelationshipFilter(ctx, s, i, filter, results)

		case models.RelHasOne:
			// these shoud contain the foreign key in their definitions
			go handleHasOneRelationshipFilter(ctx, s, i, filter, results)

		case models.RelMany2Many:
			// TODO: finish the many2many relationship
			results <- i
			wg.Done()

			log.Debugf("RelMany2Many filters not implemented yet!")
		}
	}

	var ctr int
fl:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v := <-results:
			if err, ok := v.(error); ok {
				return err
			}
			ctr++

			if ctr == len(relationshipFilters) {
				break fl
			}
		}
	}

	// if all relationship filters are done we can clear them
	(*scope.Scope)(s).ClearRelationshipFilters()

	return nil
}

func handleBelongsToRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {
	log.Debugf("SCOPE[%s] handleBelongsToRelationshipFilter", s.ID())
	// for the belongs to relationships if the relationship filters only over the relations primary
	// key then we can change the filter from relationship into the foreign key for the root scope
	// otherwise we need to get the

	var onlyPrimes = true
	for _, nested := range filter.NestedFields() {
		if nested.StructField().FieldKind() != models.KindPrimary {
			onlyPrimes = false
			break
		}
	}

	if onlyPrimes {
		// convert the filters into the foreign key filters of the root scope
		var foreignKey = filter.StructField().Relationship().ForeignKey()

		filterField := (*scope.Scope)(s).GetOrCreateForeignKeyFilter(foreignKey)

		for _, nested := range filter.NestedFields() {
			filterField.AddValues(nested.Values()...)
		}

		// send the index of the filter to remove from the scope
		results <- index
		return
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	for _, nested := range filter.NestedFields() {
		if err := (*scope.Scope)(relScope).AddFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)

			// send the error to the results
			results <- err
			return
		}
	}

	// we only need primary keys
	relScope.SetFieldset(relScope.Struct().Primary())

	if err := relScope.ListContext(ctx); err != nil {
		results <- err
		return
	}

	primaries, err := (*scope.Scope)(relScope).GetPrimaryFieldValues()
	if err != nil {
		results <- err
		return
	}

	if len(primaries) == 0 {
		log.Debugf("SCOPE[%s] getting BelongsTo relationship filters no results found.", s.ID().String())
		results <- unidb.ErrNoResult.New()
		return
	}

	// convert the filters into the foreign key filters of the root scope
	var foreignKey = filter.StructField().Relationship().ForeignKey()

	filterField := (*scope.Scope)(s).GetOrCreateForeignKeyFilter(foreignKey)

	filterField.AddValues(filters.NewOpValuePair(filters.OpIn, primaries...))

	// send the index of the filter to remove from the scope
	results <- index

	return
}

// has many relationships is a relationship where the foreign model contains the foreign key
// in order to match the given relationship the related scope must be taken with the foreign key in the fieldset
// having the foreign key from the related model, the root scope should have primary field filter key with
// the values of the related model's foreign key results
func handleHasManyRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {

	// if all the nested filters are the foreign key of this relationship then there is no need to get values from scope
	var onlyForeign = true

	foreignKey := filter.StructField().Relationship().ForeignKey()

	for _, nested := range filter.NestedFields() {
		if nested.StructField() != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// convert the foreign into root scope primary key filter
		primaryFilter := (*scope.Scope)(s).GetOrCreateIDFilter()

		for _, nested := range filter.NestedFields() {
			primaryFilter.AddValues(nested.Values()...)
		}

		results <- index
		return
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	for _, nested := range filter.NestedFields() {
		if err := (*scope.Scope)(relScope).AddFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)

			// send the error to the results
			results <- err
			return
		}
	}

	// the only required field in fieldset is a foreign key
	(*scope.Scope)(relScope).SetFields(foreignKey)

	if err := relScope.ListContext(ctx); err != nil {
		results <- err
		return
	}

	foreignValues, err := (*scope.Scope)(relScope).GetForeignKeyValues(foreignKey)
	if err != nil {
		log.Errorf("Getting ForeignKeyValues in a relationship scope failed: %v", err)
		results <- err
		return
	}

	if len(foreignValues) == 0 {
		results <- unidb.ErrNoResult.New()
		log.Debugf("No results found for the relationship filter for field: %s", filter.StructField().ApiName())
		return
	}

	primaryFilter := (*scope.Scope)(s).GetOrCreateIDFilter()
	primaryFilter.AddValues(filters.NewOpValuePair(filters.OpIn, foreignValues...))

	results <- index

}

// has many relationships is a relationship where the foreign model contains the foreign key
// in order to match the given relationship the related scope must be taken with the foreign key in the fieldset
// having the foreign key from the related model, the root scope should have primary field filter key with
// the values of the related model's foreign key results
func handleHasOneRelationshipFilter(ctx context.Context, s *Scope, index int, filter *filters.FilterField, results chan<- interface{}) {

	// if all the nested filters are the foreign key of this relationship then there is no need to get values from scope
	var onlyForeign = true

	foreignKey := filter.StructField().Relationship().ForeignKey()

	for _, nested := range filter.NestedFields() {
		if nested.StructField() != foreignKey {
			onlyForeign = false
			break
		}
	}

	if onlyForeign {
		// convert the foreign into root scope primary key filter
		primaryFilter := (*scope.Scope)(s).GetOrCreateIDFilter()

		for _, nested := range filter.NestedFields() {
			primaryFilter.AddValues(nested.Values()...)
		}

		results <- index
		return
	}

	relScope := NewModelC(s.Controller(), (*mapping.ModelStruct)(filter.StructField().Relationship().Struct()), true)

	for _, nested := range filter.NestedFields() {
		if err := (*scope.Scope)(relScope).AddFilterField(nested); err != nil {
			log.Error("Adding filter for the relation scope failed: %v", err)

			// send the error to the results
			results <- err
			return
		}
	}

	// the only required field in fieldset is a foreign key
	(*scope.Scope)(relScope).SetFields(foreignKey)

	if err := relScope.ListContext(ctx); err != nil {
		results <- err
		return
	}

	foreignValues, err := (*scope.Scope)(relScope).GetForeignKeyValues(foreignKey)
	if err != nil {
		log.Errorf("Getting ForeignKeyValues in a relationship scope failed: %v", err)
		results <- err
		return
	}

	if len(foreignValues) == 0 {
		results <- unidb.ErrNoResult.New()
		return
	}

	if len(foreignValues) > 1 && !(*scope.Scope)(s).IsMany() {
		log.Warningf("Multiple foreign key values found for a single valued - has one scope.")
	}

	primaryFilter := (*scope.Scope)(s).GetOrCreateIDFilter()
	primaryFilter.AddValues(filters.NewOpValuePair(filters.OpIn, foreignValues...))

	results <- index
}

// processGetIncluded gets the included fields for the
func getIncludedFunc(ctx context.Context, s *Scope) error {
	log.Debugf("getIncludedFunc")
	iScope := (*scope.Scope)(s)

	if iScope.IsRoot() && len(iScope.IncludedScopes()) == 0 {
		return nil
	}

	if err := iScope.SetCollectionValues(); err != nil {
		log.Debugf("SetCollectionValues for model: '%v' failed. Err: %v", s.Struct().Collection(), err)
		return err
	}

	maxTimeout := s.Controller().Config.Builder.RepositoryTimeout
	for _, incScope := range iScope.IncludedScopes() {

		if incScope.Struct().Config() == nil {
			continue
		}
		if modelRepo := incScope.Struct().Config().Repository; modelRepo != nil {
			if tm := modelRepo.MaxTimeout; tm != nil {
				if *tm > maxTimeout {
					maxTimeout = *tm
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	includedFields := iScope.IncludedFields()

	results := make(chan interface{}, len(includedFields))

	// get include job
	getInclude := func(includedField *scope.IncludeField, results chan<- interface{}) {

		missing, err := includedField.GetMissingPrimaries()
		if err != nil {
			log.Debugf("Model: %v, includedField: '%s', GetMissingPrimaries failed: %v", s.Struct().Collection(), includedField.Name(), err)
			results <- err
			return
		}
		log.Debugf("Missing primaries: %v", missing)

		if len(missing) > 0 {
			includedScope := includedField.Scope

			includedScope.SetIDFilters(missing...)
			includedScope.NewValueMany()

			if err = (*Scope)(includedScope).ListContext(ctx); err != nil {
				log.Debugf("Model: %v, includedField '%s' Scope.List failed. %v", s.Struct().Collection(), includedField.Name(), err)
				results <- err
				return
			}

		}
		results <- struct{}{}
	}

	// send the jobs
	for _, includedField := range includedFields {
		go getInclude(includedField, results)
	}

	// collect the results
	var ctr int
	for {
		select {
		case <-ctx.Done():
		case v, ok := <-results:
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				return err
			}
			ctr++

			if ctr == len(includedFields) {
				break
			}
		}
	}
}

func getForeignRelationshipsFunc(ctx context.Context, s *Scope) error {
	log.Debugf("Scope: %s, getForeignRelationship", s.ID().String())

	relations := map[string]*models.StructField{}

	// get all relationship field from the fieldset
	for _, field := range (*scope.Scope)(s).Fieldset() {
		if field.Relationship() != nil {
			if _, ok := relations[field.ApiName()]; !ok {
				relations[field.ApiName()] = field
			}
		}
	}

	// If the included is not within fieldset, get their primaries also.
	for _, included := range (*scope.Scope)(s).IncludedFields() {
		if _, ok := relations[included.ApiName()]; !ok {
			relations[included.ApiName()] = included.StructField
		}
	}

	if len(relations) > 0 {

		v := reflect.ValueOf(s.Value)
		if v.IsNil() {
			log.Debugf("Scope has a nil value.")
			return internal.ErrNilValue
		}
		v = v.Elem()

		// log.Debugf("Internal check for the ctrl: %p", ctx.Value(internal.ControllerCtxKey))

		results := make(chan interface{}, len(relations))

		// create the cancelable context for the sub context
		maxTimeout := s.Controller().Config.Builder.RepositoryTimeout
		for _, rel := range relations {
			if rel.Relationship().Struct().Config() == nil {
				continue
			}
			if modelRepo := rel.Relationship().Struct().Config().Repository; modelRepo != nil {
				if tm := modelRepo.MaxTimeout; tm != nil {
					if *tm > maxTimeout {
						maxTimeout = *tm
					}
				}
			}
		}

		ctx, cancel := context.WithTimeout(ctx, maxTimeout)
		defer cancel()

		for _, rel := range relations {
			go getForeignRelationshipValue(ctx, s, v, rel, results)
		}

		var ctr int
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case v, ok := <-results:
				if !ok {
					return nil
				}

				if err, ok := v.(error); ok {
					return err
				}

				ctr++
				if ctr == len(relations) {
					return nil
				}
			}
		}
	}
	return nil
}

func getForeignRelationshipValue(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
	res chan<- interface{},
) {

	var err error

	defer func() {
		if err != nil {
			res <- err
		} else {
			res <- struct{}{}
		}
	}()
	// Check the relationships
	rel := relField.Relationship()

	switch rel.Kind() {
	case models.RelHasOne:
		// if relationship is synced get the relations primaries from the related repository
		// if scope value is many use List
		//
		// if scope value is single use Get
		//
		// the scope should contain only foreign key as fieldset
		// Create new relation scope
		var (
			op           *filters.Operator
			filterValues []interface{}
		)

		sync := rel.Sync()
		if sync != nil && !*sync {
			return
		}

		if !(*scope.Scope)(s).IsMany() {

			// Value already checked if nil
			op = filters.OpEqual

			primVal := v.FieldByIndex(s.Struct().Primary().ReflectField().Index)
			prim := primVal.Interface()

			// check if primary field is zero
			if reflect.DeepEqual(prim, reflect.Zero(primVal.Type()).Interface()) {
				return
			}

			filterValues = []interface{}{prim}
		} else {
			op = filters.OpIn
			// Iterate over an slice of values

			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.IsNil() {
					continue
				}

				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}

				primVal := elem.FieldByIndex(s.Struct().Primary().ReflectField().Index)

				prim := primVal.Interface()
				// Check if primary field is zero valued
				if reflect.DeepEqual(prim, reflect.Zero(primVal.Type()).Interface()) {
					continue
				}

				// if not add it to the filter values
				filterValues = append(filterValues, prim)
			}

			if len(filterValues) == 0 {
				return
			}
		}

		// Create the
		fk := rel.ForeignKey()

		relatedScope := newScopeWithModel(s.Controller(), rel.Struct(), (*scope.Scope)(s).IsMany())

		// set the fieldset
		relatedScope.SetEmptyFieldset()
		relatedScope.SetFieldsetNoCheck(rel.Struct().PrimaryField(), fk)
		relatedScope.SetFlagsFrom((*scope.Scope)(s).Flags())
		relatedScope.SetCollectionScope(relatedScope)

		// set filterfield
		filter := filters.NewFilter(fk, filters.NewOpValuePair(op, filterValues...))

		relatedScope.AddFilterField(filter)

		if (*scope.Scope)(s).IsMany() {
			err = (*Scope)(relatedScope).ListContext(ctx)
			if err != nil {
				dberr, ok := err.(*unidb.Error)
				if ok {
					if dberr.Compare(unidb.ErrNoResult) {
						err = nil
						return
					}
				}

				return
			}

			// iterate over relatedScope values and match the fk's with scope's primaries.
			relVal := reflect.ValueOf(relatedScope.Value)
			if relVal.IsNil() {
				log.Errorf("Relationship field's scope has nil value after List. Rel: %s", relField.ApiName())
				err = internal.ErrNilValue
				return
			}

			relVal = relVal.Elem()
			for i := 0; i < relVal.Len(); i++ {
				elem := relVal.Index(i)

				if elem.Kind() == reflect.Ptr {
					if elem.IsNil() {
						continue
					}
					elem = elem.Elem()
				}

				// ForeignKey Field Value
				fkVal := elem.FieldByIndex(fk.ReflectField().Index)

				// Primary Field from the Related Scope Value
				pkVal := elem.FieldByIndex(relatedScope.Struct().PrimaryField().ReflectField().Index)

				pk := pkVal.Interface()

				if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
					log.Debugf("Relationship HasMany. Elem value with Zero Value for foreign key. Elem: %#v, FKName: %s", elem.Interface(), fk.ApiName())
					continue
				}

			rootLoop:
				for j := 0; j < v.Len(); j++ {
					rootElem := v.Index(j)
					if rootElem.Kind() == reflect.Ptr {
						rootElem = rootElem.Elem()
					}

					rootPrim := rootElem.FieldByIndex(s.Struct().Primary().ReflectField().Index)
					if rootPrim.Interface() == fkVal.Interface() {
						rootElem.FieldByIndex(relField.ReflectField().Index).Set(relVal.Index(i))
						break rootLoop
					}
				}

			}
		} else {

			err = (*Scope)(relatedScope).GetContext(ctx)
			if err != nil {
				dberr, ok := err.(*unidb.Error)
				if ok && dberr.Compare(unidb.ErrNoResult) {
					err = nil
					return
				}
				return
			}
			// After getting the values set the scope relation value into the value of the relatedscope
			rf := v.FieldByIndex(relField.ReflectField().Index)
			if rf.Kind() == reflect.Ptr {
				rf.Set(reflect.ValueOf(relatedScope.Value))
			} else {
				rf.Set(reflect.ValueOf(relatedScope.Value).Elem())
			}
		}

	case models.RelHasMany:

		// if relationship is synced get it from the related repository
		// Use a relatedRepo.List method
		// the scope should contain only foreign key as fieldset
		var (
			op           *filters.Operator
			filterValues []interface{}

			primMap   map[interface{}]int
			primIndex = s.Struct().Primary().ReflectField().Index
		)

		sync := rel.Sync()
		if sync != nil && !*sync {
			return
		}

		// If is single scope's value
		if !(*scope.Scope)(s).IsMany() {

			op = filters.OpEqual
			pkVal := v.FieldByIndex(primIndex)
			pk := pkVal.Interface()
			if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
				// if the primary value is not set the function should not enter here
				err = fmt.Errorf("Empty Scope Primary Value")
				log.Debugf("Err: %v. pk:%v, Scope: %#v", err, pk, s)
				return
			}

			rf := v.FieldByIndex(relField.ReflectField().Index)
			rf.Set(reflect.Zero(relField.ReflectField().Type))

			filterValues = append(filterValues, pk)
		} else {
			primMap = map[interface{}]int{}
			op = filters.OpIn
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)

				if elem.Kind() == reflect.Ptr {
					if elem.IsNil() {
						continue
					}
					elem = elem.Elem()
				}

				pkVal := elem.FieldByIndex(primIndex)
				pk := pkVal.Interface()

				// if the value is Zero like continue
				if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
					continue
				}

				primMap[pk] = i

				rf := elem.FieldByIndex(relField.ReflectField().Index)
				rf.Set(reflect.Zero(relField.ReflectField().Type))
				filterValues = append(filterValues, pk)
			}
		}
		log.Debugf("Filter Values: %v", filterValues)
		if len(filterValues) == 0 {
			return
		}
		relatedScope := newScopeWithModel(s.Controller(), rel.Struct(), true)
		relatedScope.SetCollectionScope(relatedScope)

		fk := rel.ForeignKey()

		filterField := filters.NewFilter(fk, filters.NewOpValuePair(op, filterValues...))

		if err = relatedScope.AddFilterField(filterField); err != nil {
			log.Errorf("AddingFilterField failed. %v", err)
			return
		}
		// clear the fieldset
		relatedScope.SetEmptyFieldset()

		// set the fields to primary field and foreign key
		relatedScope.SetFieldsetNoCheck(relatedScope.Struct().PrimaryField(), fk)

		if err = (*Scope)(relatedScope).ListContext(ctx); err != nil {
			dberr, ok := err.(*unidb.Error)
			if ok {
				if dberr.Compare(unidb.ErrNoResult) {
					err = nil
					return
				}
			}
			log.Debugf("Error while getting foreign field: '%s'. Err: %v", rel.BackrefernceFieldName(), err)
			return
		}

		relVal := reflect.ValueOf(relatedScope.Value).Elem()

		for i := 0; i < relVal.Len(); i++ {
			elem := relVal.Index(i)

			if elem.Kind() == reflect.Ptr {
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
			}

			// get the foreignkey value
			fkVal := elem.FieldByIndex(fk.ReflectField().Index)
			// pkVal := elem.FieldByIndex(relatedScope.Struct.primary.refStruct.Index)

			if (*scope.Scope)(s).IsMany() {

				// foreign key in relation would be a primary key within root scope
				j, ok := primMap[fkVal.Interface()]
				if !ok {

					// TODO: write some good debug log
					log.Debugf("Foreign Key not found in the primaries map.")
					continue
				}

				// rootPK should be at index 'j'
				var scopeElem reflect.Value
				scopeElem = v.Index(j)
				if scopeElem.Kind() == reflect.Ptr {
					scopeElem = scopeElem.Elem()
				}

				elemRelVal := scopeElem.FieldByIndex(relField.ReflectField().Index)
				elemRelVal.Set(reflect.Append(elemRelVal, relVal.Index(i)))
			} else {
				elemRelVal := v.FieldByIndex(relField.ReflectField().Index)
				elemRelVal.Set(reflect.Append(elemRelVal, relVal.Index(i)))
			}

		}

	case models.RelMany2Many:
		// if the relationship is synced get the values from the relationship
		sync := rel.Sync()
		if sync != nil && *sync {
			// when sync get the relationships value from the backreference relationship
			var (
				filterValues = []interface{}{}
				primMap      map[interface{}]int
				primIndex    = s.Struct().Primary().ReflectField().Index
				op           *filters.Operator
			)

			if !(*scope.Scope)(s).IsMany() {
				// set isMany to false just to see the difference

				op = filters.OpEqual

				pkVal := v.FieldByIndex(primIndex)
				pk := pkVal.Interface()

				if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
					err = fmt.Errorf("Error no primary valoue set for the scope value.  Scope %#v. Value: %+v", s, s.Value)
					return
				}

				filterValues = append(filterValues, pk)

			} else {

				// Operator is 'in'
				op = filters.OpIn

				// primMap contains the index of the proper scope value for
				//given primary key value
				primMap = map[interface{}]int{}
				for i := 0; i < v.Len(); i++ {
					elem := v.Index(i)

					if elem.Kind() == reflect.Ptr {
						if elem.IsNil() {
							continue
						}
						elem = elem.Elem()
					}

					pkVal := elem.FieldByIndex(primIndex)
					pk := pkVal.Interface()

					// if the value is Zero like continue
					if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
						continue
					}

					primMap[pk] = i

					filterValues = append(filterValues, pk)
				}
			}

			backReference := rel.BackreferenceField()

			backRefRelPrimary := backReference.Relationship().Struct().PrimaryField()

			// prepare new scope
			relatedScope := newScopeWithModel(s.Controller(), rel.Struct(), true)

			relatedScope.SetEmptyFieldset()
			relatedScope.SetFieldsetNoCheck(backReference)
			relatedScope.SetCollectionScope(relatedScope)

			// Add filter
			filterField := filters.NewFilter(backReference, filters.NewOpValuePair(op, filterValues...))
			relatedScope.AddFilterField(filterField)

			if err = (*Scope)(relatedScope).ListContext(ctx); err != nil {
				dberr, ok := err.(*unidb.Error)
				if ok {
					if dberr.Compare(unidb.ErrNoResult) {
						err = nil
						return
					}
				}
				return
			}

			relVal := reflect.ValueOf(relatedScope.Value)
			if relVal.IsNil() {
				log.Errorf("Related Field scope: %s is nil after getting m2m relationships.", relField.ApiName())
				err = internal.ErrNilValue
				return
			}

			relVal = relVal.Elem()

			// iterate over relatedScoep Value
			for i := 0; i < relVal.Len(); i++ {
				// Get Each element from the relatedScope Value
				relElem := relVal.Index(i)

				// Dereference the ptr
				if relElem.Kind() == reflect.Ptr {
					if relElem.IsNil() {
						continue
					}
					relElem = relElem.Elem()
				}

				relPrimVal := relElem.FieldByIndex(relatedScope.Struct().PrimaryField().ReflectField().Index)

				// continue if empty
				if reflect.DeepEqual(relPrimVal.Interface(), reflect.Zero(relPrimVal.Type()).Interface()) {
					continue
				}

				backRefVal := relElem.FieldByIndex(backReference.ReflectField().Index)
				if backRefVal.Kind() == reflect.Ptr {
					if backRefVal.IsNil() {
						backRefVal.Set(reflect.New(backRefVal.Type().Elem()))
					}
				}
				backRefVal.Elem()

				// iterate over backreference elems
			backRefLoop:
				for j := 0; j < backRefVal.Len(); j++ {
					// BackRefPrimary would be root primary
					backRefElem := backRefVal.Index(j)

					if backRefElem.Kind() == reflect.Ptr {
						if backRefElem.IsNil() {
							continue
						}

						backRefElem = backRefElem.Elem()
					}

					backRefPrimVal := backRefElem.FieldByIndex(backRefRelPrimary.ReflectField().Index)

					backRefPrim := backRefPrimVal.Interface()

					if reflect.DeepEqual(backRefPrim, reflect.Zero(backRefPrimVal.Type()).Interface()) {
						log.Warningf("Backreference Relationship Many2Many contains zero valued primary key. Scope: %#v, RelatedScope: %#v. Relationship: %#v", s, relatedScope, rel)
						continue backRefLoop
					}

					var val reflect.Value
					if (*scope.Scope)(s).IsMany() {
						index, ok := primMap[backRefPrim]
						if !ok {
							log.Warningf("Relationship Many2Many contains backreference primaries that doesn't match the root scope value. Scope: %#v. Relationship: %v", s, rel)
							continue backRefLoop
						}

						val = v.Index(index)
					} else {
						val = v
					}

					m2mElem := reflect.New(rel.Struct().Type())
					m2mElem.Elem().FieldByIndex(rel.Struct().PrimaryField().ReflectField().Index).Set(relPrimVal)

					relationField := val.FieldByIndex(relField.ReflectField().Index)
					relationField.Set(reflect.Append(relationField, m2mElem))
				}

			}

		}
	case models.RelBelongsTo:

		// if scope value is a slice
		// iterate over all entries and transfer all foreign keys into relationship primaries.

		// for single value scope do it once

		sync := rel.Sync()
		if sync != nil && *sync {
			// don't know what to do now
			// Get it from the backreference ? and
			log.Warningf("Synced BelongsTo relationship. Scope: %#v, Relation: %#v", s, rel)
		}

		fkField := rel.ForeignKey()

		if !(*scope.Scope)(s).IsMany() {

			fkVal := v.FieldByIndex(fkField.ReflectField().Index)

			// check if foreign key is zero
			if reflect.DeepEqual(fkVal.Interface(), reflect.Zero(fkVal.Type()).Interface()) {

				return
			}

			relVal := v.FieldByIndex(relField.ReflectField().Index)

			if relVal.Kind() == reflect.Ptr {
				if relVal.IsNil() {
					relVal.Set(reflect.New(relVal.Type().Elem()))
				}
				relVal = relVal.Elem()
			} else if relVal.Kind() != reflect.Struct {
				err = fmt.Errorf("Relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", s.Struct().Collection(), rel)
				log.Warning(err)
				return
			}

			relPrim := relVal.FieldByIndex(rel.Struct().PrimaryField().ReflectField().Index)
			relPrim.Set(fkVal)

		} else {
			// is a list type scope
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.Kind() == reflect.Ptr {
					if elem.IsNil() {
						continue
					}
					elem = elem.Elem()
				}

				// check if foreign key is not a zero
				fkVal := elem.FieldByIndex(fkField.ReflectField().Index)
				if reflect.DeepEqual(fkVal.Interface(), reflect.Zero(fkVal.Type()).Interface()) {
					continue
				}

				relVal := elem.FieldByIndex(relField.ReflectField().Index)
				if relVal.Kind() == reflect.Ptr {
					if relVal.IsNil() {
						relVal.Set(reflect.New(relVal.Type().Elem()))
					}
					relVal = relVal.Elem()
				} else if relVal.Kind() != reflect.Struct {
					err = fmt.Errorf("Relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", s.Struct().Collection(), rel)
					log.Warning(err)
					return
				}
				relPrim := relVal.FieldByIndex(rel.Struct().PrimaryField().ReflectField().Index)
				relPrim.Set(fkVal)
			}
		}
	}

}

// setBelongsToRelationshipsFunc sets the value from the belongs to relationship ID's to the foreign keys
func setBelongsToRelationshipsFunc(ctx context.Context, s *Scope) error {
	return (*scope.Scope)(s).SetBelongsToForeignKeyFields()
}
