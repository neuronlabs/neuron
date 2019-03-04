package internal

import (
	"fmt"
)

func SplitBracketParameter(bracketed string) (values []string, err error) {
	// look for values in
	doubleOpen := func() error {
		return fmt.Errorf("Open square bracket '[' found, without closing ']' in: '%s'.",
			bracketed)
	}

	var startIndex int = -1
	var endIndex int = -1
	for i := 0; i < len(bracketed); i++ {
		c := bracketed[i]
		switch c {
		case AnnotationOpenedBracket:
			if startIndex > endIndex {
				err = doubleOpen()
				return
			}
			startIndex = i
		case AnnotationClosedBracket:
			// if opening bracket not set or in case of more than one brackets
			// if start was not set before this endIndex
			if startIndex == -1 || startIndex < endIndex {
				err = fmt.Errorf("Close square bracket ']' found, without opening '[' in '%s'.", bracketed)
				return
			}
			endIndex = i
			values = append(values, bracketed[startIndex+1:endIndex])
		}
	}
	if (startIndex != -1 && endIndex == -1) || startIndex > endIndex {
		err = doubleOpen()
		return
	}
	return
}
