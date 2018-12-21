package dbmanager

import (
	"errors"
	aerrors "github.com/kucjac/jsonapi/errors"
	"github.com/kucjac/uni-db"
	"sync"
)

// DefaultErrorMap contain default mapping of unidb.Error prototype into
// Error. It is used by default by 'ErrorManager' if created using New() function.
var DefaultErrorMap map[unidb.Error]aerrors.ApiError = map[unidb.Error]aerrors.ApiError{
	unidb.ErrNoResult:              aerrors.ErrResourceNotFound,
	unidb.ErrConnExc:               aerrors.ErrInternalError,
	unidb.ErrCardinalityViolation:  aerrors.ErrInternalError,
	unidb.ErrDataException:         aerrors.ErrInvalidInput,
	unidb.ErrIntegrConstViolation:  aerrors.ErrInvalidInput,
	unidb.ErrRestrictViolation:     aerrors.ErrInvalidInput,
	unidb.ErrNotNullViolation:      aerrors.ErrInvalidInput,
	unidb.ErrForeignKeyViolation:   aerrors.ErrInvalidInput,
	unidb.ErrUniqueViolation:       aerrors.ErrResourceAlreadyExists,
	unidb.ErrCheckViolation:        aerrors.ErrInvalidInput,
	unidb.ErrInvalidTransState:     aerrors.ErrInternalError,
	unidb.ErrInvalidTransTerm:      aerrors.ErrInternalError,
	unidb.ErrTransRollback:         aerrors.ErrInternalError,
	unidb.ErrTxDone:                aerrors.ErrInternalError,
	unidb.ErrInvalidAuthorization:  aerrors.ErrInsufficientAccPerm,
	unidb.ErrInvalidPassword:       aerrors.ErrInternalError,
	unidb.ErrInvalidSchemaName:     aerrors.ErrInternalError,
	unidb.ErrInvalidSyntax:         aerrors.ErrInternalError,
	unidb.ErrInsufficientPrivilege: aerrors.ErrInsufficientAccPerm,
	unidb.ErrInsufficientResources: aerrors.ErrInternalError,
	unidb.ErrProgramLimitExceeded:  aerrors.ErrInternalError,
	unidb.ErrSystemError:           aerrors.ErrInternalError,
	unidb.ErrInternalError:         aerrors.ErrInternalError,
	unidb.ErrUnspecifiedError:      aerrors.ErrInternalError,
}

// ErrorManager defines the database unidb.Error one-to-one mapping
// into Error. The default error mapping is defined
// in package variable 'DefaultErrorMap'.
//
type ErrorManager struct {
	dbToRest map[unidb.Error]aerrors.ApiError
	sync.RWMutex
}

// NewErrorMapper creates new error handler with already inited ErrorMap
func NewDBErrorMgr() *ErrorManager {
	return &ErrorManager{dbToRest: DefaultErrorMap}
}

// Handle enables unidb.Error handling so that proper ErrorObject is returned.
// It returns ErrorObject if given database error exists in the private error mapping.
// If provided dberror doesn't have prototype or no mapping exists for given unidb.Error an
// application 'error' would be returned.
// Thread safety by using RWMutex.RLock
func (r *ErrorManager) Handle(dberr *unidb.Error) (*aerrors.ApiError, error) {
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

// LoadCustomErrorMap enables replacement of the ErrorManager default error map.
// This operation is thread safe - with RWMutex.Lock
func (r *ErrorManager) LoadCustomErrorMap(errorMap map[unidb.Error]aerrors.ApiError) {
	r.Lock()
	r.dbToRest = errorMap
	r.Unlock()
}

// UpdateErrorMapEntry changes single entry in the Error Handler error map.
// This operation is thread safe - with RWMutex.Lock
func (r *ErrorManager) UpdateErrorEntry(
	dberr unidb.Error,
	apierr aerrors.ApiError,
) {
	r.Lock()
	r.dbToRest[dberr] = apierr
	r.Unlock()
}
