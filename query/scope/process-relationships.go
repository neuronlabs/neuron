package scope

import (
	"context"
	"github.com/kucjac/uni-db"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	scope "github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/pkg/errors"
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
)

// processGetIncluded gets the included fields for the
func getIncludedFunc(s *Scope) error {
	log.Debugf("getIncludedFunc")
	iScope := (*scope.Scope)(s)

	if iScope.IsRoot() && len(iScope.IncludedScopes()) == 0 {
		return nil
	}

	if err := iScope.SetCollectionValues(); err != nil {
		log.Debugf("SetCollectionValues for model: '%v' failed. Err: %v", s.Struct().Collection(), err)
		return err
	}

	for iScope.NextIncludedField() {
		includedField, err := iScope.CurrentIncludedField()
		if err != nil {
			return err
		}
		log.Debugf("Getting Included for the field: %v", includedField.ApiName())

		missing, err := includedField.GetMissingPrimaries()
		if err != nil {
			log.Debugf("Model: %v, includedField: '%s', GetMissingPrimaries failed: %v", s.Struct().Collection(), includedField.Name(), err)
			return err
		}
		log.Debugf("Missing primaries: %v", missing)

		if len(missing) > 0 {
			includedField.Scope.SetIDFilters(missing...)
			includedField.Scope.NewValueMany()

			if err := DefaultQueryProcessor.DoList((*Scope)(includedField.Scope)); err != nil {
				log.Debugf("Model: %v, includedField '%s' Scope.List failed. %v", s.Struct().Collection(), includedField.Name(), err)
				return err
			}
		}
	}
	iScope.ResetIncludedField()
	return nil
}

func getForeignRelationshipsFunc(s *Scope) error {
	log.Debugf("getForeignRelationship")
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

	v := reflect.ValueOf(s.Value)
	if v.IsNil() {
		log.Debugf("Scope has a nil value.")
		return internal.IErrNilValue
	}
	v = v.Elem()

	ctx := s.Context()

	// log.Debugf("Internal check for the ctrl: %p", ctx.Value(internal.ControllerKeyCtxKey))

	for _, rel := range relations {
		err := getForeignRelationshipValue(ctx, s, v, rel)
		if err != nil {
			return err
		}
	}

	return nil
}

func getForeignRelationshipValue(
	ctx context.Context,
	s *Scope,
	v reflect.Value,
	relField *models.StructField,
) error {
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
			return nil
		}

		if !(*scope.Scope)(s).IsMany() {

			// Value already checked if nil
			op = filters.OpEqual

			primVal := v.FieldByIndex(s.Struct().Primary().ReflectField().Index)
			prim := primVal.Interface()

			// check if primary field is zero
			if reflect.DeepEqual(prim, reflect.Zero(primVal.Type()).Interface()) {
				return nil
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
				return nil
			}
		}

		// Create the
		fk := rel.ForeignKey()
		relatedScope := scope.NewWithCtx(ctx, rel.Struct())

		relatedScope.SetEmptyFieldset()
		relatedScope.SetFieldsetNoCheck(fk)
		relatedScope.SetFlagsFrom((*scope.Scope)(s).Flags())
		relatedScope.SetCollectionScope(relatedScope)

		// set filterfield
		filter := filters.NewFilter(fk, filters.NewOpValuePair(op, filterValues...))

		relatedScope.AddFilterField(filter)

		if (*scope.Scope)(s).IsMany() {
			relatedScope.NewValueMany()
			err := (*Scope)(relatedScope).List()
			if err != nil {
				dberr, ok := err.(*unidb.Error)
				if ok {
					if dberr.Compare(unidb.ErrNoResult) {
						return nil
					}
				}
				return err
			}

			// iterate over relatedScope values and match the fk's with scope's primaries.
			relVal := reflect.ValueOf(relatedScope.Value)
			if relVal.IsNil() {
				log.Errorf("Relationship field's scope has nil value after List. Rel: %s", relField.ApiName())
				return internal.IErrNilValue
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
			relatedScope.NewValueSingle()
			err := (*Scope)(relatedScope).Get()
			if err != nil {
				dberr, ok := err.(*unidb.Error)
				if ok && dberr.Compare(unidb.ErrNoResult) {
					return nil
				}
				return err
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
			return nil
		}

		// If is single scope's value
		if !(*scope.Scope)(s).IsMany() {

			op = filters.OpEqual
			pkVal := v.FieldByIndex(primIndex)
			pk := pkVal.Interface()
			if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
				// if the primary value is not set the function should not enter here
				err := errors.Errorf("Empty Scope Primary Value")
				log.Debugf("Err: %v. pk:%v, Scope: %#v", err, pk, s)
				return err
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

		relatedScope := scope.NewWithCtx(ctx, rel.Struct())
		relatedScope.SetCollectionScope(relatedScope)

		fk := rel.ForeignKey()

		filterField := filters.NewFilter(fk, filters.NewOpValuePair(op, filterValues...))

		if err := relatedScope.AddFilterField(filterField); err != nil {
			panic(err)
		}
		relatedScope.SetEmptyFieldset()
		relatedScope.SetFieldsetNoCheck(fk)

		// set fieldset
		relatedScope.NewValueMany()

		if err := (*Scope)(relatedScope).List(); err != nil {
			dberr, ok := err.(*unidb.Error)
			if ok {
				if dberr.Compare(unidb.ErrNoResult) {
					return nil
				}
			}
			log.Debugf("Error while getting foreign field: '%s'. Err: %v", rel.BackrefernceFieldName(), err)
			return err
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
					err := errors.Errorf("Error no primary valoue set for the scope value.  Scope %#v. Value: %+v", s, s.Value)
					return err
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
			relatedScope := scope.NewWithCtx(ctx, rel.Struct())
			relatedScope.SetEmptyFieldset()
			relatedScope.SetFieldsetNoCheck(backReference)
			relatedScope.SetCollectionScope(relatedScope)

			// Add filter

			filterField := filters.NewFilter(backReference, filters.NewOpValuePair(op, filterValues...))
			relatedScope.AddFilterField(filterField)

			relatedScope.NewValueMany()

			if err := (*Scope)(relatedScope).List(); err != nil {
				dberr, ok := err.(*unidb.Error)
				if ok {
					if dberr.Compare(unidb.ErrNoResult) {
						return nil
					}
				}
				return err
			}

			relVal := reflect.ValueOf(relatedScope.Value)
			if relVal.IsNil() {
				log.Errorf("Related Field scope: %s is nil after getting m2m relationships.", relField.ApiName())
				return internal.IErrNilValue
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
				return nil
			}

			relVal := v.FieldByIndex(relField.ReflectField().Index)

			if relVal.Kind() == reflect.Ptr {
				if relVal.IsNil() {
					relVal.Set(reflect.New(relVal.Type().Elem()))
				}
				relVal = relVal.Elem()
			} else if relVal.Kind() != reflect.Struct {
				err := errors.Errorf("Relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", s.Struct().Collection(), rel)
				log.Warning(err)
				return err
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
					err := errors.Errorf("Relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", s.Struct().Collection(), rel)
					log.Warning(err)
					return err
				}
				relPrim := relVal.FieldByIndex(rel.Struct().PrimaryField().ReflectField().Index)
				relPrim.Set(fkVal)
			}
		}
	}
	return nil
}

func setBelongsToRelationshipsFunc(s *Scope) error {
	return (*scope.Scope)(s).SetBelongsToForeignKeyFields()
}
