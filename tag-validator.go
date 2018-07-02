package jsonapi

import (
	"reflect"
	"strings"
)

func JSONAPITagFunc(field reflect.StructField) string {
	tagValue, ok := field.Tag.Lookup("jsonapi")
	if !ok || tagValue == "" {
		return ""
	}

	splitted := strings.Split(tagValue, ",")
	if len(splitted) > 2 {
		for _, v := range splitted[1:] {
			if v == "hidden" {
				return ""
			}
		}
	} else if len(splitted) == 1 {
		return ""
	}

	switch splitted[0] {
	case "primary":
		return "id"
	case "relation":
		return splitted[1]
	case "attr":
		return splitted[1]
	}
	return ""

}
