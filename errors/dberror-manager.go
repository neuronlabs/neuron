package errors

import (
	"errors"
	"github.com/kucjac/uni-db"
	"sync"
)

// DefaultErrorMap contain default mapping of unidb.Error prototype into
// Error. It is used by default by 'ErrorMapper' if created using New() function.
var DefaultErrorMap map[unidb.Error]ApiError = map[unidb.Error]ApiError{
	unidb.ErrNoResult:              ErrResourceNotFound,
	unidb.ErrConnExc:               ErrInternalError,
	unidb.ErrCardinalityViolation:  ErrInternalError,
	unidb.ErrDataException:         ErrInvalidInput,
	unidb.ErrIntegrConstViolation:  ErrInvalidInput,
	unidb.ErrRestrictViolation:     ErrInvalidInput,
	unidb.ErrNotNullViolation:      ErrInvalidInput,
	unidb.ErrForeignKeyViolation:   ErrInvalidInput,
	unidb.ErrUniqueViolation:       ErrResourceAlreadyExists,
	unidb.ErrCheckViolation:        ErrInvalidInput,
	unidb.ErrInvalidTransState:     ErrInternalError,
	unidb.ErrInvalidTransTerm:      ErrInternalError,
	unidb.ErrTransRollback:         ErrInternalError,
	unidb.ErrTxDone:                ErrInternalError,
	unidb.ErrInvalidAuthorization:  ErrInsufficientAccPerm,
	unidb.ErrInvalidPassword:       ErrInternalError,
	unidb.ErrInvalidSchemaName:     ErrInternalError,
	unidb.ErrInvalidSyntax:         ErrInternalError,
	unidb.ErrInsufficientPrivilege: ErrInsufficientAccPerm,
	unidb.ErrInsufficientResources: ErrInternalError,
	unidb.ErrProgramLimitExceeded:  ErrInternalError,
	unidb.ErrSystemError:           ErrInternalError,
	unidb.ErrInternalError:         ErrInternalError,
	unidb.ErrUnspecifiedError:      ErrInternalError,
}

// ErrorMapper defines the database unidb.Error one-to-one mapping
// into neuron APIError. The default error mapping is defined
// in package variable 'DefaultErrorMap'.
type ErrorMapper struct {
	dbToRest map[unidb.Error]ApiError
	sync.RWMutex
}

// NewDBMapper creates new error handler with already inited ErrorMap
func NewDBMapper() *ErrorMapper {
	return &ErrorMapper{dbToRest: DefaultErrorMap}
}

// Handle enables unidb.Error handling so that proper ErrorObject is returned.
// It returns ErrorObject if given database error exists in the private error mapping.
// If provided dberror doesn't have prototype or no mapping exists for given unidb.Error an
// application 'error' would be returned.
// Thread safety by using RWMutex.RLock
func (r *ErrorMapper) Handle(dberr *unidb.Error) (*ApiError, error) {
	// Get the prototype for given dberr
	dbProto, err := dberr.GetPrototype()
	if err != nil {
		return nil, err
	}

	// Get Rest
	r.RLock()
	apierr, ok := r.dbToRest[dbProto]
	r.RUnlock()
	if !ok {
		err = errors.New("Given database error is unrecognised by the handler")
		return nil, err
	}

	// // Create new entity
	return &apierr, nil
}

// LoadCustomErrorMap enables replacement of the ErrorMapper default error map.
// This operation is thread safe - with RWMutex.Lock
func (r *ErrorMapper) LoadCustomErrorMap(errorMap map[unidb.Error]ApiError) {
	r.Lock()
	r.dbToRest = errorMap
	r.Unlock()
}

// UpdateErrorEntry changes single entry in the Error Handler error map.
// This operation is thread safe - with RWMutex.Lock
func (r *ErrorMapper) UpdateErrorEntry(
	dberr unidb.Error,
	apierr ApiError,
) {
	r.Lock()
	r.dbToRest[dberr] = apierr
	r.Unlock()
}
