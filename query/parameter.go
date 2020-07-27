package query

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/neuronlabs/neuron/errors"
)

const intSize = 32 << (^uint(0) >> 63)

// MakeParameters creates new parameters from provided url values.
func MakeParameters(values url.Values) (parameters Parameters) {
	for k, v := range values {
		parameters = append(parameters, Parameter{Key: k, Value: v[0]})
	}
	return parameters
}

// Parameters is the slice wrapper for the query parameters.
type Parameters []Parameter

// Get gets the parameter stored as 'key'.
func (p Parameters) Get(key string) (Parameter, bool) {
	for i := 0; i < len(p); i++ {
		if p[i].Key == key {
			return p[i], true
		}
	}
	return Parameter{}, false
}

// Exists checks if given 'key' exists in the parameters
func (p Parameters) Exists(key string) bool {
	for i := 0; i < len(p); i++ {
		if p[i].Key == key {
			return true
		}
	}
	return false
}

// Set parameter with 'key' and 'value'.
func (p *Parameters) Set(key, value string) {
	*p = append(*p, Parameter{Key: key, Value: value})
}

// SetInt sets the integer parameter value.
func (p *Parameters) SetInt(key string, i int) {
	*p = append(*p, Parameter{Key: key, Value: strconv.Itoa(i)})
}

// SetInt64 sets the 64 bit integer value.
func (p *Parameters) SetInt64(key string, i int64) {
	*p = append(*p, Parameter{Key: key, Value: strconv.FormatInt(i, 10)})
}

// SetFloat64 sets the 64-bit float value and trims to selected precision.
func (p *Parameters) SetFloat64(key string, f float64, precision int) {
	*p = append(*p, Parameter{Key: key, Value: strconv.FormatFloat(f, 'f', precision, 64)})
}

// SetUint sets the unsigned integer value.
func (p *Parameters) SetUint(key string, i uint) {
	*p = append(*p, Parameter{Key: key, Value: strconv.FormatUint(uint64(i), intSize)})
}

// SetUint64 sets the unsigned 64-bit integer value.
func (p *Parameters) SetUint64(key string, i uint64) {
	*p = append(*p, Parameter{Key: key, Value: strconv.FormatUint(i, intSize)})
}

// SetUUID sets uuid value.
func (p *Parameters) SetUUID(key string, id uuid.UUID) {
	*p = append(*p, Parameter{Key: key, Value: id.String()})
}

// SetBoolean sets the boolean value.
func (p *Parameters) SetBoolean(key string, b bool) {
	*p = append(*p, Parameter{Key: key, Value: strconv.FormatBool(b)})
}

// Parameter defines generic query key-value parameter.
type Parameter struct {
	Key, Value string
}

// Int returns integer parameter.
func (p Parameter) Int() (int, error) {
	v, err := strconv.Atoi(p.Value)
	if err != nil {
		return 0, errors.NewDetf(ClassInvalidParameter, "parameter: '%s' is not an integer: '%s'", p.Key, p.Value)
	}
	return v, nil
}

// Int64 returns 64 bit integer parameter.
func (p Parameter) Int64() (int64, error) {
	v, err := strconv.ParseInt(p.Value, 10, intSize)
	if err != nil {
		return 0, errors.NewDetf(ClassInvalidParameter, "parameter: '%s' is not an integer: '%s'", p.Key, p.Value)
	}
	return v, nil
}

// Float64 returns 64-bit float parameter.
func (p Parameter) Float64() (float64, error) {
	v, err := strconv.ParseFloat(p.Value, 64)
	if err != nil {
		return 0, errors.NewDetf(ClassInvalidParameter, "parameter: '%s' is not a floating point value: '%s'", p.Key, p.Value)
	}
	return v, nil
}

// Uint returns unsigned integer parameter.
func (p Parameter) Uint() (uint, error) {
	u, err := strconv.ParseUint(p.Value, 10, intSize)
	if err != nil {
		return 0, errors.NewDetf(ClassInvalidParameter, "parameter: '%s' is not an integer: '%s'", p.Key, p.Value)
	}
	return uint(u), err
}

// Uint64 returns unsigned 64-bit integer parameter.
func (p Parameter) Uint64() (uint64, error) {
	v, err := strconv.ParseUint(p.Value, 10, intSize)
	if err != nil {
		return 0, errors.NewDetf(ClassInvalidParameter, "parameter: '%s' is not an integer: '%s'", p.Key, p.Value)
	}
	return v, nil
}

// UUID returns uuid integer parameter.
func (p Parameter) UUID() (uuid.UUID, error) {
	v, err := uuid.Parse(p.Value)
	if err != nil {
		return uuid.UUID{}, errors.NewDetf(ClassInvalidParameter, "parameter: '%s' is not a valid UUID: '%s'", p.Key, p.Value)
	}
	return v, nil
}

// Boolean return boolean parameter value.
func (p Parameter) Boolean() (bool, error) {
	v, err := strconv.ParseBool(p.Value)
	if err != nil {
		return false, errors.NewDetf(ClassInvalidParameter, "parameter: '%s' is not a valid boolean: '%s'", p.Key, p.Value)
	}
	return v, nil
}

// StringSlice returns string slice parameter values.
func (p Parameter) StringSlice() []string {
	return strings.Split(p.Value, ",")
}
