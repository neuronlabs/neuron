package annotation

// Relation strategy relationship fields tags.
const (
	OnCreate = "on_create"
	OnPatch  = "on_patch"
	OnDelete = "on_delete"
	Order    = "order"
	OnError  = "on_error"
	OnChange = "on_change"

	RelationRestrict = "restrict"
	RelationNoAction = "no-action"
	RelationCascade  = "cascade"
	RelationSetNull  = "set-null"

	FailOnError     = "fail"
	ContinueOnError = "continue"

	Default = "default"
)
