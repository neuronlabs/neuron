package config

const (
	// FlUseLinks is the Flag that allows to return query links
	FlUseLinks uint = 1 << iota
	FlReturnPatchContent
	FlAddMetaCountList
	FlAllowClientID

	// AllowForeignKeyFilter is the flag that allows filtering over foreign keys
	FlAllowForeignKeyFilter

	// UseFilterValueLimit is the flag that checks if there is any limit for the filter values
	FlUseFilterValueLimit

	// AllowStringSearch is a flag that defines if the string field may be filtered using
	// operators like: '$contains', '$startswith', '$endswith'
	FlAllowStringSearch
)
