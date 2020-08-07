package server

import (
	"net/http"
)

// Middleware is a http middleware function.
type Middleware func(http.Handler) http.Handler

// MiddlewareChain is a middleware slice type that could be used as a http.Handler.
type MiddlewareChain []Middleware

// Handle handles the chain of middlewares for given 'handle' http.Handle.
func (c MiddlewareChain) Handle(handle http.Handler) http.Handler {
	for i := range c {
		handle = c[len(c)-1-i](handle)
	}
	return handle
}
