package config

import (
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/log"
	"github.com/pkg/errors"
	"strconv"
)

const (
	// FlUseLinks is the Flag that allows to return query links
	FlUseLinks uint = iota

	// FlReturnPatchContent is a flag that sets the behaviour of returning the patch content for the request
	FlReturnPatchContent

	// FlAddMetaCountList is a flag that adds the count list to the meta as a default behavior
	FlAddMetaCountList

	// FlAllowClientID is a flag that allows to use the Client defined ID
	FlAllowClientID

	// FlAllowForeignKeyFilter is the flag that allows filtering over foreign keys
	FlAllowForeignKeyFilter

	// FlUseFilterValueLimit is the flag that checks if there is any limit for the filter values
	FlUseFilterValueLimit

	// FlAllowStringSearch is a flag that defines if the string field may be filtered using
	// operators like: '$contains', '$startswith', '$endswith'
	FlAllowStringSearch
)

// Flags defines flag configurations
type Flags map[string]interface{}

// Container returns flags container
func (f *Flags) Container() (*flags.Container, error) {
	c := flags.New()
	for k, v := range *f {
		var (
			value bool
			err   error
		)
		switch b := v.(type) {
		case string:
			value, err = strconv.ParseBool(b)
			if err != nil {
				return nil, errors.Wrapf(err, "Invalid flag value: %s", b)
			}
		case bool:
			value = b
		case int:
			if b > 0 {
				value = true
			}
		case float64:
			if b > 0 {
				value = true
			}
		default:
			return nil, errors.Wrapf(err, "Invalid flag value: %v", b)
		}

		switch k {
		case "return_links":
			c.Set(FlUseLinks, value)
		case "return_patch_content":
			c.Set(FlReturnPatchContent, value)
		case "add_meta_count_list":
			c.Set(FlAddMetaCountList, value)
		case "allow_client_id":
			c.Set(FlAllowClientID, value)
		case "allow_foreign_key_filters":
			c.Set(FlAllowForeignKeyFilter, value)
		case "use_filter_values_limit":
			c.Set(FlUseFilterValueLimit, value)
		case "allow_string_search":
			c.Set(FlAllowStringSearch, value)
		default:
			log.Infof("Provided invalid key: '%s' for flag definition.", k)
		}
	}
	return c, nil
}
