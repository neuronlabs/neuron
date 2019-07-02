package filters

import (
	"strconv"
	"time"

	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
)

func stringValue(sField *mapping.StructField, val interface{}, values *[]string) {
	switch v := val.(type) {
	case string:
		*values = append(*values, v)
	case uint:
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case uint8:
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case uint16:
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case uint32:
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case uint64:
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case int:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int8:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int16:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int32:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int64:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case float32:
		*values = append(*values, strconv.FormatFloat(float64(v), 'f', -1, 32))
	case float64:
		*values = append(*values, strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		*values = append(*values, strconv.FormatBool(v))
	case QueryValuer:
		*values = append(*values, v.QueryValue(sField))
	case time.Time:
		*values = append(*values, strconv.FormatInt(v.Unix(), 10))
	case *time.Time:
		*values = append(*values, strconv.FormatInt(v.Unix(), 10))
	case []interface{}:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []time.Time:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []*time.Time:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []int64:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []int32:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []int16:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []int8:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []uint8:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []uint16:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []uint32:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []uint64:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []uint:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []QueryValuer:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	case []string:
		for _, s := range v {
			stringValue(sField, s, values)
		}
	default:
		log.Debugf("Filter parse query. Invalid or unsupported value type: %t", v)
	}

}
