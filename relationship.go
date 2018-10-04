package jsonapi

import (
	"github.com/pkg/errors"
	"reflect"
)

type RelationshipKind int

const (
	RelUnknown RelationshipKind = iota
	RelBelongsTo
	RelHasOne
	RelHasMany
	RelMany2Many
	RelMany2ManyDisjoint
	RelMany2ManyCommon
)

// By default the relationship many2many is local
// it can be synced with back-referenced relationship using the 'sync' flag
// The HasOne and HasMany is the relationship that by default is 'synced'
// The BelongsTo relationship is local by default.

type Relationship struct {
	// Kind is a relationship kind
	Kind RelationshipKind

	// ForeignKey represtents the foreignkey field
	ForeignKey *StructField

	// Sync is a flag that defines if the relationship opertaions
	// should be synced with the related many2many relationship
	// or the foreignkey in related model
	Sync *bool

	// BackReference Fieldname is a field name that is back-reference
	// relationship in many2many relationships
	BackReferenceFieldname string

	// BackReferenceField
	BackReferenceField *StructField
}

func (r Relationship) isToOne() bool {
	switch r.Kind {
	case RelHasOne, RelBelongsTo:
		return true
	}
	return false
}

func (r Relationship) isToMany() bool {
	switch r.Kind {
	case RelHasOne, RelBelongsTo, RelUnknown:
		return false
	}
	return true
}

func (r Relationship) isMany2Many() bool {
	switch r.Kind {
	case RelMany2ManyCommon, RelMany2ManyDisjoint, RelMany2Many:
		return true
	}
	return false
}

func (c *Controller) setRelationships() error {
	for _, model := range c.Models.models {
		for _, relField := range model.relationships {

			if relField.relationship == nil {
				relField.relationship = &Relationship{}
			}
			relationship := relField.relationship

			// get structfield jsonapi tags
			tags, err := relField.getTagValues(relField.refStruct.Tag.Get(annotationJSONAPI))
			if err != nil {
				return err
			}

			// get proper foreign key field name
			fkeyFieldName := tags.Get(annotationForeignKey)

			// check field type
			switch relField.refStruct.Type.Kind() {
			case reflect.Slice:
				// has many by default
				if relationship.isMany2Many() {
					if relationship.Sync != nil && !(*relationship.Sync) {
						continue
					}
					if bf := relationship.BackReferenceFieldname; bf != "" {
						bf = getNameByConvention(bf, c.NamingStrategy)
						backReferenced, ok := relField.relatedStruct.relationships[bf]
						if !ok {
							err = errors.Errorf("The backreference collection named: '%s' is invalid. Model: %s, Sfield: '%s'", bf, model.modelType.Name(), relField.refStruct.Name)
							return err
						}
						relationship.BackReferenceField = backReferenced
					}
					continue
				}

				// HasMany
				relationship.Kind = RelHasMany

				if relationship.Sync != nil && !(*relationship.Sync) {
					continue
				}

				if fkeyFieldName == "" {
					fkeyFieldName = model.modelType.Name() + "ID"
				}
				fkeyName := getNameByConvention(fkeyFieldName, c.NamingStrategy)
				fk, ok := relField.relatedStruct.foreignKeys[fkeyName]
				if !ok {
					return errors.Errorf("Foreign key not found for the relationship: '%s'. Model: '%s'", relField.fieldName, model.modelType.Name())
				}

				relationship.ForeignKey = fk
				b := true

				relationship.Sync = &b

			case reflect.Ptr, reflect.Struct:
				// check if it is belongs_to or has_one relationship
				// at first search for foreign key as
				if fkeyFieldName == "" {
					fkeyFieldName = relField.refStruct.Name + "ID"
				}
				fkeyName := getNameByConvention(fkeyFieldName, c.NamingStrategy)

				nosync := (relationship.Sync != nil && !*relationship.Sync)
				fk, ok := model.foreignKeys[fkeyName]
				if !ok {
					c.log().Debugf("Not found for %s", fkeyName)
					relationship.Kind = RelHasOne
					if nosync {
						continue
					}
					fk, ok = relField.relatedStruct.foreignKeys[fkeyName]
					if !ok {
						return errors.Errorf("Foreign key not found for the relationship: '%s'. Model: '%s'", relField.fieldName, model.modelType.Name())
					}

				} else {
					c.log().Debugf("found for: %s", relField.fieldName)
					relationship.Kind = RelBelongsTo
				}
				relationship.ForeignKey = fk
			}
		}
	}
	return nil
}
