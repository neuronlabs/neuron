package filters

// Constants that define operator raw values.
const (
	AnnotationOperatorEqual        = "$eq"
	AnnotationOperatorIn           = "$in"
	AnnotationOperatorNotEqual     = "$ne"
	AnnotationOperatorNotIn        = "$notin"
	AnnotationOperatorGreaterThan  = "$gt"
	AnnotationOperatorGreaterEqual = "$ge"
	AnnotationOperatorLessThan     = "$lt"
	AnnotationOperatorLessEqual    = "$le"
	AnnotationOperatorIsNull       = "$isnull"
	AnnotationOperatorNotNull      = "$notnull"
	AnnotationOperatorExists       = "$exists"
	AnnotationOperatorNotExists    = "$notexists"
	AnnotationOperatorContains     = "$contains"
	AnnotationOperatorStartsWith   = "$startswith"
	AnnotationOperatorEndsWith     = "$endswith"
	AnnotationOperatorStDWithin    = "$st_dwithin"
)
