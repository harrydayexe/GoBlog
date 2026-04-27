package config

import "github.com/harrydayexe/GoWebUtilities/middleware"

// BaseServerOption represents a configuration option for the HTTP server.
// Options use the functional options pattern; each value carries a function
// pointer that modifies a specific server setting.
//
// This type should not be constructed directly. Use the provided option
// functions: WithPort, WithHost, and WithMiddleware.
type BaseServerOption struct {
	BaseOption

	WithPortFunc       func(v *Port)
	WithHostFunc       func(v *Host)
	WithMiddlewareFunc func(mw *[]middleware.Middleware)
}

// Port is the TCP port number the HTTP server listens on.
type Port int

// WithPort returns a BaseServerOption that sets the HTTP server listen port.
//
// Example usage:
//
//	cfg := config.ServerConfig{
//	    Server: []config.BaseServerOption{
//	        config.WithPort(8080),
//	    },
//	}
func WithPort(port int) BaseServerOption {
	return BaseServerOption{
		WithPortFunc: func(v *Port) { *v = Port(port) },
	}
}

// Host is the network address the HTTP server binds to.
// An empty Host binds to all available network interfaces.
type Host string

// WithHost returns a BaseServerOption that sets the HTTP server bind address.
//
// Example usage:
//
//	cfg := config.ServerConfig{
//	    Server: []config.BaseServerOption{
//	        config.WithHost("127.0.0.1"),
//	    },
//	}
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
