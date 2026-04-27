package config

import "github.com/harrydayexe/GoWebUtilities/middleware"

type BaseServerOption struct {
	BaseOption

	WithPortFunc       func(v *Port)
	WithHostFunc       func(v *Host)
	WithMiddlewareFunc func(mw *[]middleware.Middleware)
}

type Port int

func WithPort(port int) BaseServerOption {
	return BaseServerOption{
		WithPortFunc: func(v *Port) { *v = Port(port) },
	}
}

type Host string

func WithHost(host string) BaseServerOption {
	return BaseServerOption{
		WithHostFunc: func(v *Host) { *v = Host(host) },
	}
}

// WithMiddleware returns a BaseServerOption that adds HTTP middleware to the server.
//
// Middleware are applied in the order provided. The first middleware in the list
// will be the outermost wrapper (executed first for requests, last for responses).
// Multiple calls to WithMiddleware append to the middleware chain.
//
// Example usage:
//
//	import (
//	    "github.com/harrydayexe/GoWebUtilities/logging"
//	    "github.com/harrydayexe/GoWebUtilities/middleware"
//	)
//
//	cfg := config.ServerConfig{
//	    Server: []config.BaseServerOption{
//	        config.WithPort(8080),
//	        config.WithMiddleware(
//	            logging.New(logger),      // Built-in logging
//	            customAuthMiddleware,     // Custom middleware
//	        ),
//	    },
//	}
//
// All middleware must be safe for concurrent use by multiple goroutines.
func WithMiddleware(mw ...middleware.Middleware) BaseServerOption {
	return BaseServerOption{
		WithMiddlewareFunc: func(middleware *[]middleware.Middleware) {
			*middleware = append(*middleware, mw...)
		},
	}
}
