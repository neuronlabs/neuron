package query

// Operators
const (
	// logical filters
	operatorEqual        = "$eq"
	operatorIn           = "$in"
	operatorNotEqual     = "$ne"
	operatorNotIn        = "$notin"
	operatorGreaterThan  = "$gt"
	operatorGreaterEqual = "$ge"
	operatorLessThan     = "$lt"
	operatorLessEqual    = "$le"
	operatorIsNull       = "$isnull"
	operatorNotNull      = "$notnull"
	operatorExists       = "$exists"
	operatorNotExists    = "$notexists"
	operatorContains     = "$contains"
	operatorStartsWith   = "$startswith"
	operatorEndsWith     = "$endswith"
	operatorStDWithin    = "$st_dwithin"
)
