package jsonapi

import (
	"testing"
)

func TestSetRelationScopeSort(t *testing.T) {
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)
	mStruct := c.MustGetModelStruct(&Blog{})

	sortScope := &SortScope{Field: mStruct.GetPrimaryField()}
	inv := sortScope.setRelationScope([]string{}, AscendingOrder)
	assertTrue(t, inv)

	sortScope = &SortScope{Field: mStruct.relationships["posts"]}
	inv = sortScope.setRelationScope([]string{}, AscendingOrder)
	assertTrue(t, inv)

	inv = sortScope.setRelationScope([]string{"posts", "some", "id"}, AscendingOrder)
	assertTrue(t, inv)

	inv = sortScope.setRelationScope([]string{"comments", "id", "desc"}, AscendingOrder)
	assertTrue(t, inv)

	inv = sortScope.setRelationScope([]string{"comments", "id"}, AscendingOrder)
	assertFalse(t, inv)

	inv = sortScope.setRelationScope([]string{"comments", "body"}, AscendingOrder)
	assertFalse(t, inv)

	inv = sortScope.setRelationScope([]string{"comments", "id"}, AscendingOrder)
	assertFalse(t, inv)

}
