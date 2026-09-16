// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"time"

	"github.com/harrydayexe/GoWebUtilities/middleware"
)

// BaseServerOption represents a configuration option for the HTTP server.
// Options use the functional options pattern; each value carries a function
// pointer that modifies a specific server setting.
//
// This type should not be constructed directly. Use the provided option
// functions: [WithPort], [WithHost], [WithMiddleware], [WithCacheControl],
// [WithHealthChecks], or call [BaseOption.AsServerOption] on a [BaseOption]
// value (e.g. from [WithLogger]).
type BaseServerOption struct {
	BaseOption

	WithPortFunc         func(v *Port)
	WithHostFunc         func(v *Host)
	WithMiddlewareFunc   func(mw *[]middleware.Middleware)
	WithCacheControlFunc func(v *CacheControlTTL)
	WithHealthChecksFunc func(v *HealthChecks)
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

// AsServerOption returns a BaseServerOption that applies this BaseOption to a
// server instance, enabling a BaseOption (e.g. from [WithLogger] or
// [WithBlogRoot]) to be passed to server constructors alongside other
// server options.
func (o BaseOption) AsServerOption() BaseServerOption {
	return BaseServerOption{
		BaseOption: o,
	}
}

// CacheControlTTL holds the max-age duration for the Cache-Control response
// header. When TTL is greater than zero the server adds
// "Cache-Control: public, max-age=<seconds>" to every response, telling
// browsers and shared caches how long they may serve the response without
// revalidating. A TTL of zero or negative disables the header entirely.
//
// The default TTL (applied when no [WithCacheControl] option is supplied) is
// one hour.
type CacheControlTTL struct{ TTL time.Duration }

// WithCacheControl returns a BaseServerOption that sets the Cache-Control
// max-age TTL on all HTTP responses served by the server.
//
// When ttl > 0 the server adds "Cache-Control: public, max-age=<N>" to every
// response, where N is ttl truncated to whole seconds. Setting ttl to 0 (or
// any non-positive value) disables the header so no Cache-Control is sent.
//
// The default (when this option is not supplied) is one hour.
//
// Example usage:
//
//	cfg := config.ServerConfig{
//	    Server: []config.BaseServerOption{
//	        config.WithCacheControl(24 * time.Hour), // cache for one day
//	    },
//	}
//
// To disable caching entirely:
//
//	cfg.Server = append(cfg.Server, config.WithCacheControl(0))
func WithCacheControl(ttl time.Duration) BaseServerOption {
	return BaseServerOption{
		WithCacheControlFunc: func(v *CacheControlTTL) { v.TTL = ttl },
	}
}

// HealthChecks is a configuration type that controls whether the HTTP server
// exposes health-check endpoints at /healthz/live, /healthz/ready, and
// /healthz/startup.
//
// When Enabled is true, the server binds the HTTP listener before loading posts
// and templates, so liveness/readiness/startup probes can reach the endpoints
// during startup. Health checks are disabled by default; the Docker image
// enables them via the --health-checks flag.
//
// This type is typically embedded in the server struct and set via [WithHealthChecks].
type HealthChecks struct{ Enabled bool }

// WithHealthChecks returns a [BaseServerOption] that enables health-check
// endpoints on the HTTP server.
//
// When enabled, the server exposes three unauthenticated endpoints:
//   - GET /healthz/live    — always 200 OK (process is alive)
//   - GET /healthz/ready   — 200 OK once posts and templates have loaded;
//     503 Service Unavailable while starting up or if loading failed
//   - GET /healthz/startup — same semantics as /healthz/ready
//
// The server binds the HTTP listener before initialising content when health
// checks are enabled, allowing probes to observe the startup state. Response
// bodies are plain text ("ok" or an error description).
//
// Health checks are disabled by default.
//
// Example usage:
//
//	cfg := config.ServerConfig{
//	    Server: []config.BaseServerOption{
//	        config.WithHealthChecks(),
//	    },
//	}
func WithHealthChecks() BaseServerOption {
	return BaseServerOption{
		WithHealthChecksFunc: func(v *HealthChecks) {
			v.Enabled = true
		},
	}
}
