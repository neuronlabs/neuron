package jsonapi

// Error is the jsonapi error structure. It is used for marshaling and unmarshaling purpose.
// More info can be found at: 'https://jsonapi.org/format/#errors'
type Error struct {
	// ID is a unique identifier for this particular occurrence of a problem.
	ID string `json:"id,omitempty"`
	// Title is a short, human-readable summary of the problem that SHOULD NOT change from occurrence to occurrence of the problem, except for purposes of localization.
	Title string `json:"title,omitempty"`
	// Detail is a human-readable explanation specific to this occurrence of the problem. Like title, this field’s value can be localized.
	Detail string `json:"detail,omitempty"`
	// Status is the HTTP status code applicable to this problem, expressed as a string value.
	Status string `json:"status,omitempty"`
	// Code is an application-specific error code, expressed as a string value.
	Code string `json:"code,omitempty"`
	// Meta is an object containing non-standard meta-information about the error.
	Meta map[string]interface{} `json:"meta,omitempty"`
}
