package query

//
// // TestScopeSorts tests the scope sort functions.
// func TestScopeSorts(t *testing.T) {
// 	type SortForeignModel struct {
// 		ID     int
// 		Length int
// 	}
//
// 	type SortModel struct {
// 		ID         int
// 		Name       string
// 		Age        int
// 		Relation   *SortForeignModel
// 		RelationID int
// 	}
//
// 	c := newController(t)
//
// 	err := c.RegisterModels(SortModel{}, SortForeignModel{})
// 	require.NoError(t, err)
//
// 	t.Run("EmptyFields", func(t *testing.T) {
// 		s := NewScopeC(c, &SortModel{})
// 		require.NoError(t, err)
//
// 		err = s.SortField("-TransactionID")
// 		require.NoError(t, err)
//
// 		err = s.Sort("-age")
// 		assert.NoError(t, err)
// 	})
//
// 	t.Run("Duplicated", func(t *testing.T) {
// 		s := NewScopeC(c, &SortModel{})
// 		require.NoError(t, err)
//
// 		err = s.Sort("age")
// 		require.NoError(t, err)
//
// 		err = s.Sort("-TransactionID", "TransactionID", "-TransactionID")
// 		assert.Error(t, err)
// 	})
// }
