// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// defaultMetricsPort is the port the admin listener binds to when
// --metrics-port is not supplied.
const defaultMetricsPort = 9090

// metricsPath is the only route the admin listener serves.
const metricsPath = "/metrics"

// metricsShutdownTimeout matches the graceful shutdown budget
// [github.com/harrydayexe/GoBlog/v2/pkg/server.Server.Run] gives the blog
// listener, so both listeners drain under the same deadline.
const metricsShutdownTimeout = 10 * time.Second

// metricsServer owns the OpenTelemetry pipeline behind --metrics and the admin
// HTTP listener that exposes it.
//
// The admin listener is deliberately separate from the blog's listener:
// operational data stays off the public port, the route is unaffected by
// --blog-root prefixing, and the port can be left unpublished in a container.
type metricsServer struct {
	provider *sdkmetric.MeterProvider
	listener net.Listener
	server   *http.Server
}

// newMetricsServer builds a Prometheus-backed meter provider and binds the
// admin listener on host:port. An empty host binds to all interfaces.
//
// The listener is bound here rather than inside [metricsServer.Serve] so that a
// port clash fails the serve command before the blog starts: serving a blog
// without the metrics an operator believes are running is worse than failing
// loudly.
//
// The exporter registers with a private registry rather than the default
// Prometheus one, so the admin listener exposes exactly the instruments the
// blog records.
func newMetricsServer(host string, port int) (*metricsServer, error) {
	registry := promclient.NewRegistry()
	exporter, err := prometheus.New(prometheus.WithRegisterer(registry))
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		_ = provider.Shutdown(context.Background())
		return nil, fmt.Errorf("failed to bind metrics listener on %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET "+metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))

	return &metricsServer{
		provider: provider,
		listener: listener,
		server:   &http.Server{Handler: mux},
	}, nil
}

// MeterProvider returns the provider the server records its instruments
// against, for passing to config.WithMeterProvider.
func (m *metricsServer) MeterProvider() *sdkmetric.MeterProvider { return m.provider }

// Addr returns the address the admin listener is bound to. It is resolved from
// the listener, so a --metrics-port of 0 reports the port the OS chose.
func (m *metricsServer) Addr() string { return m.listener.Addr().String() }

// Serve answers /metrics until the server is shut down. It returns nil on a
// clean shutdown.
func (m *metricsServer) Serve() error {
	if err := m.server.Serve(m.listener); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown stops the admin listener gracefully and then flushes the meter
// provider, so measurements taken up to the final request are still readable by
// a scrape in flight.
func (m *metricsServer) Shutdown(ctx context.Context) error {
	serverErr := m.server.Shutdown(ctx)
	// Shutdown only closes the listeners Serve has already registered, so a
	// server stopped before Serve ran would leave the port held until the
	// goroutine caught up. Closing it here makes the release deterministic; the
	// second close Serve performs is a no-op error we do not care about.
	_ = m.listener.Close()

	providerErr := m.provider.Shutdown(ctx)

	if serverErr != nil {
		return fmt.Errorf("failed to shut down metrics listener: %w", serverErr)
	}
	if providerErr != nil {
		return fmt.Errorf("failed to shut down meter provider: %w", providerErr)
	}
	return nil
}
