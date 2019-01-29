package middlewares

import (
	"errors"
	"net/http"
)

var (
	ErrMiddlewareAlreadyRegistered error = errors.New("Middleware already registered")
	ErrMiddlewareNotRegistered     error = errors.New("Middleware not registered.")
)

type MiddlewareFunc func(next http.Handler) http.Handler

// Middleware defines the middleware function with it's name to use in the endpoints
type middleware struct {
	Name string
	Func func(next http.Handler) http.Handler
}
