package gateway

type HandlerError struct {
	Code    ErrorCode
	Message string
	Scope   *Scope
	Field   *StructField
	Model   *ModelStruct
}

func newHandlerError(code ErrorCode, msg string) *HandlerError {
	return &HandlerError{Code: code, Message: msg}
}

func (e *HandlerError) Error() string {
	return fmt.Sprintf("%d. %s", e.Code, e.Message)
}
