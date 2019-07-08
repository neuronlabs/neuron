package mapping

import (
	"strings"
)

// FieldTag is the key: values pair for the given field struct's tag.
type FieldTag struct {
	Key    string
	Values []string
}

// ExtractFieldTags extracts the []*mapping.FieldTag from the given *mapping.StructField
// for given StructField reflect tag.
func (s *StructField) ExtractFieldTags(fieldTag string) []*FieldTag {
	tag, ok := s.ReflectField().Tag.Lookup(fieldTag)
	if !ok {
		// if there is no struct tag with name 'fieldTag' return nil
		return nil
	}

	// omit the field with the '-' tag
	if tag == "-" {
		return []*FieldTag{{Key: "-"}}
	}

	var (
		seperators []int
		tags       []*FieldTag
		options    []string
	)
	// find all the seperators
	for i, r := range tag {
		if i != 0 && r == ';' {
			// check if the  rune before is not an 'escape'
			if tag[i-1] != '\\' {
				seperators = append(seperators, i)
			}
		}
	}

	for i, sep := range seperators {
		if i == 0 {
			options = append(options, tag[:sep])
		} else {
			options = append(options, tag[seperators[i-1]+1:sep])
		}

		if i == len(seperators)-1 {
			options = append(options, tag[sep+1:])
		}
	}

	// if no seperators found add the option as whole tag tag
	if options == nil {
		options = append(options, tag)
	}

	// options should be now a legal values defined for the struct tag

	for _, o := range options {
		var equalIndex int

		// find the equalIndex
		for i, r := range o {
			if r == '=' {
				if i != 0 && o[i-1] != '\\' {
					equalIndex = i
					break
				}
			}
		}

		tag := &FieldTag{}

		if equalIndex != 0 {
			// the left part would be the key
			tag.Key = o[:equalIndex]

			// the right would be the values
			tag.Values = strings.Split(o[equalIndex+1:], ",")
		} else {
			// in that case only the key should exists
			tag.Key = o
		}
		tags = append(tags, tag)

	}
	return tags
}
