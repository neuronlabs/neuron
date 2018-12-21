package query

import (
	"testing"
)

func TestSetRelationScopeSort(t *testing.T) {
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)
	mStruct := c.MustGetModelStruct(&Blog{})

	sortScope := &SortField{StructField: mStruct.GetPrimaryField()}
	inv := sortScope.setSubfield([]string{}, AscendingOrder)
	assertTrue(t, inv)

	sortScope = &SortField{StructField: mStruct.relationships["posts"]}
	inv = sortScope.setSubfield([]string{}, AscendingOrder)
	assertTrue(t, inv)

	inv = sortScope.setSubfield([]string{"posts", "some", "id"}, AscendingOrder)
	assertTrue(t, inv)

	inv = sortScope.setSubfield([]string{"comments", "id", "desc"}, AscendingOrder)
	assertTrue(t, inv)

	inv = sortScope.setSubfield([]string{"comments", "id"}, AscendingOrder)
	assertFalse(t, inv)

	inv = sortScope.setSubfield([]string{"comments", "body"}, AscendingOrder)
	assertFalse(t, inv)

	inv = sortScope.setSubfield([]string{"comments", "id"}, AscendingOrder)
	assertFalse(t, inv)

}
