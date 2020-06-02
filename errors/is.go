package errors

// IsClass checks if given error is of given 'class'.
func IsClass(err error, class Class) bool {
	classError, ok := err.(ClassError)
	if !ok {
		return false
	}
	return classError.Class() == class
}
