package jsonapi

import (
	ctrl "github.com/kucjac/jsonapi/controller"
	ictrl "github.com/kucjac/jsonapi/internal/controller"
	"io"
)

// UnmarshalRegistered unmarshals the incoming reader stream into provided
// assuming that the value 'v' is already registered in the default controller
func UnmarshalRegistered(r io.Reader, v interface{}) error {
	return (*ictrl.Controller)(ctrl.Default()).Unmarshal(r, v)
}

// UnmarshalRegisteredC unmarshals the incoming reader stream 'r' into provided model
// 'v' assuming that it is already registered within the controller 'c'
func UnmarshalRegisteredC(c *ctrl.Controller, r io.Reader, v interface{}) error {
	return nil
}

// Unmarshal incoming read input
func Unmarshal(r io.Reader, v interface{}) error {
	return nil
}

func UnmarshalC(c *ctrl.Controller, r io.Reader, v interface{}) error {
	return nil
}
