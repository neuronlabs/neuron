package errors

// ClassError is the interface used for all errors
// that uses classification system.
type ClassError interface {
	error
	// Class gets current error classification.
	Class() Class
}
