package jsonapi

import (
	"github.com/kucjac/jsonapi/ctrl"
	"io"
)

func Marshal(w io.Writer, v interface{}) error {
	err := c.Marshal(w, v)
}
