package proxy

import (
	"context"
)

// Proxy is the repository that is a proxy for the queries and passes them up to the .
type Proxy interface {
	ProxyServiceName(ctx context.Context) string
}
