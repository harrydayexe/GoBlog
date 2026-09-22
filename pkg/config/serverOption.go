// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"io/fs"
	"time"

	"github.com/harrydayexe/GoWebUtilities/middleware"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// ServerOption represents a configuration option for the HTTP server.
// Options use the functional options pattern; each value carries a function
// pointer that modifies a specific server setting.
//
// ServerOption values are passed directly to
// [github.com/harrydayexe/GoBlog/v2/pkg/server.New], which applies them to a
// [ServerConfig].
//
// This type should not be constructed directly. Use the provided option
// functions: [WithPort], [WithHost], [WithMiddleware], [WithCacheControl],
// [WithHealthChecks], [WithMeterProvider], [WithTemplateDir], or convert an
// option of another kind with [BaseOption.AsServerOption] (e.g. from
// [WithLogger]), [GeneratorOption.AsServerOption] (e.g. from [WithSiteTitle]),
// or [RendererOption.AsServerOption] (e.g. from [WithFuncs]).
type ServerOption struct {
	BaseOption

	WithPortFunc            func(v *Port)
	WithHostFunc            func(v *Host)
	WithMiddlewareFunc      func(mw *[]middleware.Middleware)
	WithCacheControlFunc    func(v *CacheControlTTL)
	WithHealthChecksFunc    func(v *HealthChecks)
	WithMeterProviderFunc   func(v *MeterProvider)
	WithTemplateDirFunc     func(v *TemplateDir)
	WithGeneratorOptionFunc func(v *[]GeneratorOption)
	WithRendererOptionFunc  func(v *[]RendererOption)
}

// Port is the TCP port number the HTTP server listens on.
type Port int

// WithPort returns a ServerOption that sets the HTTP server listen port.
//
// Example usage:
//
//	srv, err := server.New(postsFS, config.WithPort(8080))
func WithPort(port int) ServerOption {
	return ServerOption{
		WithPortFunc: func(v *Port) { *v = Port(port) },
	}
}

// Host is the network address the HTTP server binds to.
// An empty Host binds to all available network interfaces.
type Host string

// WithHost returns a ServerOption that sets the HTTP server bind address.
//
// Example usage:
//
//	srv, err := server.New(postsFS, config.WithHost("127.0.0.1"))
func WithHost(host string) ServerOption {
	return ServerOption{
		WithHostFunc: func(v *Host) { *v = Host(host) },
	}
}

// WithMiddleware returns a ServerOption that adds HTTP middleware to the server.
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
//	srv, err := server.New(postsFS,
//	    config.WithPort(8080),
//	    config.WithMiddleware(
//	        logging.New(logger),      // Built-in logging
//	        customAuthMiddleware,     // Custom middleware
//	    ),
//	)
//
// All middleware must be safe for concurrent use by multiple goroutines.
func WithMiddleware(mw ...middleware.Middleware) ServerOption {
	return ServerOption{
		WithMiddlewareFunc: func(middleware *[]middleware.Middleware) {
			*middleware = append(*middleware, mw...)
		},
	}
}

// AsServerOption returns a ServerOption that applies this BaseOption to a
// server instance, enabling a BaseOption (e.g. from [WithLogger] or
// [WithBlogRoot]) to be passed to server constructors alongside other
// server options.
func (o BaseOption) AsServerOption() ServerOption {
	return ServerOption{
		BaseOption: o,
	}
}

// AsServerOption returns a ServerOption that forwards this GeneratorOption to
// the generator the server builds internally, enabling generator options
// (e.g. from [WithSiteTitle] or [WithBaseURL]) to be passed to server
// constructors alongside other server options.
//
// Example usage:
//
//	srv, err := server.New(postsFS,
//	    config.WithPort(8080),
//	    config.WithSiteTitle("My Blog").AsServerOption(),
//	)
func (o GeneratorOption) AsServerOption() ServerOption {
	return ServerOption{
		WithGeneratorOptionFunc: func(v *[]GeneratorOption) { *v = append(*v, o) },
	}
}

// AsServerOption returns a ServerOption that forwards this RendererOption to
// the template renderer the server builds internally, enabling custom template
// functions registered with [WithFuncs] to be supplied without constructing a
// renderer manually.
//
// Example usage:
//
//	srv, err := server.New(postsFS,
//	    config.WithFuncs(template.FuncMap{"upper": strings.ToUpper}).AsServerOption(),
//	)
func (o RendererOption) AsServerOption() ServerOption {
	return ServerOption{
		WithRendererOptionFunc: func(v *[]RendererOption) { *v = append(*v, o) },
	}
}

// TemplateDir is a configuration type that holds the filesystem the HTTP
// server loads its page templates from.
//
// When FS is nil the built-in templates from
// [github.com/harrydayexe/GoBlog/v2/pkg/templates] are used.
//
// This type is typically embedded in server configuration structs and should
// be set using the [WithTemplateDir] option function.
type TemplateDir struct{ FS fs.FS }

// WithTemplateDir returns a ServerOption that sets the filesystem the server
// loads its page templates from.
//
// The filesystem must contain the same template names as
// [github.com/harrydayexe/GoBlog/v2/pkg/templates].Default. Passing nil keeps
// the built-in templates.
//
// Example usage:
//
//	srv, err := server.New(postsFS, config.WithTemplateDir(os.DirFS("templates/")))
func WithTemplateDir(fsys fs.FS) ServerOption {
	return ServerOption{
		WithTemplateDirFunc: func(v *TemplateDir) { v.FS = fsys },
	}
}

// AsOption returns a ServerOption that re-applies this TemplateDir value to
// another server.
func (o TemplateDir) AsOption() ServerOption {
	return WithTemplateDir(o.FS)
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

// WithCacheControl returns a ServerOption that sets the Cache-Control
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
//	// cache for one day
//	srv, err := server.New(postsFS, config.WithCacheControl(24*time.Hour))
//
// To disable caching entirely:
//
//	srv, err := server.New(postsFS, config.WithCacheControl(0))
func WithCacheControl(ttl time.Duration) ServerOption {
	return ServerOption{
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

// WithHealthChecks returns a [ServerOption] that enables health-check
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
//	srv, err := server.New(postsFS, config.WithHealthChecks())
func WithHealthChecks() ServerOption {
	return ServerOption{
		WithHealthChecksFunc: func(v *HealthChecks) {
			v.Enabled = true
		},
	}
}

// MeterProvider holds the OpenTelemetry meter provider the HTTP server records
// its request metrics against.
//
// The default (applied when no [WithMeterProvider] option is supplied) is the
// no-op provider from go.opentelemetry.io/otel/metric/noop, which records
// nothing. The server does not fall back to [go.opentelemetry.io/otel.GetMeterProvider],
// so instrumentation is off until a provider is passed explicitly.
//
// This type is typically embedded in the server struct and set via
// [WithMeterProvider].
type MeterProvider struct{ MeterProvider metric.MeterProvider }

// WithMeterProvider returns a [ServerOption] that sets the OpenTelemetry
// meter provider used to record HTTP server metrics.
//
// GoBlog depends on the OpenTelemetry metrics API only; the SDK and any
// exporter (Prometheus, OTLP, …) are the caller's dependency. Three
// instruments are recorded, following the stable HTTP server semantic
// conventions:
//
//   - http.server.request.duration  (histogram, seconds)
//   - http.server.active_requests   (up/down counter, requests)
//   - http.server.response.body.size (histogram, bytes)
//
// Metrics are off by default. Passing nil, or omitting the option, leaves the
// no-op provider in place and installs no instrumentation, so requests pay no
// cost. Health-check requests to /healthz/* are never recorded.
//
// Example usage, exporting to Prometheus:
//
//	import (
//	    "go.opentelemetry.io/otel/exporters/prometheus"
//	    sdkmetric "go.opentelemetry.io/otel/sdk/metric"
//	)
//
//	exporter, err := prometheus.New()
//	if err != nil {
//	    return err
//	}
//	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
//
//	srv, err := server.New(postsFS, config.WithMeterProvider(provider))
//
// To use the globally registered provider instead, pass it explicitly:
//
//	srv, err := server.New(postsFS, config.WithMeterProvider(otel.GetMeterProvider()))
func WithMeterProvider(mp metric.MeterProvider) ServerOption {
	return ServerOption{
		WithMeterProviderFunc: func(v *MeterProvider) { v.MeterProvider = mp },
	}
}

// Enabled reports whether a provider that can actually record measurements has
// been supplied. A nil provider, or the no-op provider used as the default,
// reports false so callers can skip instrumentation entirely.
func (o MeterProvider) Enabled() bool {
	if o.MeterProvider == nil {
		return false
	}
	_, isNoop := o.MeterProvider.(noop.MeterProvider)
	return !isNoop
}
