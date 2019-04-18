package flags

const (
	UseLinks uint = 1 << iota
	ReturnPatchContent
	AddMetaCountList
	AllowClientID

	// AllowForeignKeyFilter is the flag that allows filtering over foreign keys
	AllowForeignKeyFilter

	// UseFilterValueLimit is the flag that checks if there is any limit for the filter values
	UseFilterValueLimit

	// AllowStringSearch is a flag that defines if the string field may be filtered using
	// operators like: '$contains', '$startswith', '$endswith'
	AllowStringSearch
)
