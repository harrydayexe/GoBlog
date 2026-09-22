// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

const (
	durationMetric     = "http.server.request.duration"
	activeMetric       = "http.server.active_requests"
	responseSizeMetric = "http.server.response.body.size"
)

// metricsPostsFS is a posts filesystem with one tagged post, so the tag routes
// are registered and the route attribute can be asserted for them.
func metricsPostsFS() fstest.MapFS {
	return fstest.MapFS{
		"test-post.md": &fstest.MapFile{
			Data: []byte(strings.TrimSpace(`
---
title: Test Post
description: A test post for metrics testing
date: 2024-01-01
tags: [go]
---

# Test Content
			`)),
		},
	}
}

// newMetricsServer returns a server wired to a manual reader, plus a collect
// function returning the metrics recorded so far.
func newMetricsServer(t *testing.T, opts ...config.BaseServerOption) (*server.Server, func(t *testing.T) []metricdata.Metrics) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("provider.Shutdown: %v", err)
		}
	})

	cfg := config.ServerConfig{
		Server: append([]config.BaseServerOption{
			config.WithMeterProvider(provider),
			config.WithLogger(slog.New(slog.DiscardHandler)).AsServerOption(),
		}, opts...),
	}

	srv, err := server.New(nil, metricsPostsFS(), cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	return srv, func(t *testing.T) []metricdata.Metrics {
		t.Helper()
		return collect(t, reader)
	}
}

// collect drains reader and flattens every scope into one slice of metrics.
func collect(t *testing.T, reader *sdkmetric.ManualReader) []metricdata.Metrics {
	t.Helper()

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("reader.Collect: %v", err)
	}

	var out []metricdata.Metrics
	for _, sm := range rm.ScopeMetrics {
		out = append(out, sm.Metrics...)
	}
	return out
}

// findMetric returns the metric with the given name, failing the test if absent.
func findMetric(t *testing.T, ms []metricdata.Metrics, name string) metricdata.Metrics {
	t.Helper()
	for _, m := range ms {
		if m.Name == name {
			return m
		}
	}
	t.Fatalf("metric %q not recorded; got %v", name, metricNames(ms))
	return metricdata.Metrics{}
}

func metricNames(ms []metricdata.Metrics) []string {
	names := make([]string, 0, len(ms))
	for _, m := range ms {
		names = append(names, m.Name)
	}
	return names
}

// durationPoints returns the histogram data points of the request duration
// instrument.
func durationPoints(t *testing.T, ms []metricdata.Metrics) []metricdata.HistogramDataPoint[float64] {
	t.Helper()
	m := findMetric(t, ms, durationMetric)
	hist, ok := m.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatalf("%s: data is %T, want Histogram[float64]", durationMetric, m.Data)
	}
	return hist.DataPoints
}

// attrs renders an attribute set as a map for readable comparisons.
func attrs(set attribute.Set) map[string]string {
	out := make(map[string]string, set.Len())
	for _, kv := range set.ToSlice() {
		out[string(kv.Key)] = kv.Value.Emit()
	}
	return out
}

// TestMetrics_InstrumentsRecorded verifies that all three semconv instruments
// are recorded with the documented names and units.
func TestMetrics_InstrumentsRecorded(t *testing.T) {
	t.Parallel()

	srv, collected := newMetricsServer(t)

	if w := get(srv, "/"); w.Code != http.StatusOK {
		t.Fatalf("GET /: status %d, want 200", w.Code)
	}

	ms := collected(t)

	wantUnits := map[string]string{
		durationMetric:     "s",
		activeMetric:       "{request}",
		responseSizeMetric: "By",
	}
	for name, unit := range wantUnits {
		if got := findMetric(t, ms, name).Unit; got != unit {
			t.Errorf("%s unit: got %q, want %q", name, got, unit)
		}
	}

	points := durationPoints(t, ms)
	if len(points) != 1 {
		t.Fatalf("%s: got %d data points, want 1", durationMetric, len(points))
	}
	if points[0].Count != 1 {
		t.Errorf("%s: got count %d, want 1", durationMetric, points[0].Count)
	}
}

// TestMetrics_Attributes verifies the attributes recorded for each kind of
// route, including that the http.route label is the matched mux pattern.
func TestMetrics_Attributes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target string
		want   map[string]string
	}{
		{
			name:   "index",
			target: "/",
			want: map[string]string{
				"http.request.method":       "GET",
				"url.scheme":                "http",
				"http.response.status_code": "200",
				"http.route":                "GET /{$}",
			},
		},
		{
			name:   "post hit",
			target: "/posts/test-post",
			want: map[string]string{
				"http.request.method":       "GET",
				"url.scheme":                "http",
				"http.response.status_code": "200",
				"http.route":                "GET /posts/{postName}",
			},
		},
		{
			name:   "post miss",
			target: "/posts/nope",
			want: map[string]string{
				"http.request.method":       "GET",
				"url.scheme":                "http",
				"http.response.status_code": "404",
				"http.route":                "GET /posts/{postName}",
			},
		},
		{
			name:   "tag page",
			target: "/tags/go",
			want: map[string]string{
				"http.request.method":       "GET",
				"url.scheme":                "http",
				"http.response.status_code": "200",
				"http.route":                "GET /tags/{tagName}",
			},
		},
		{
			name:   "feed",
			target: "/rss.xml",
			want: map[string]string{
				"http.request.method":       "GET",
				"url.scheme":                "http",
				"http.response.status_code": "404",
				"http.route":                "GET /rss.xml",
			},
		},
		{
			name:   "asset",
			target: "/images/pipeline.png",
			want: map[string]string{
				"http.request.method":       "GET",
				"url.scheme":                "http",
				"http.response.status_code": "200",
				"http.route":                "GET /images/",
			},
		},
		{
			name:   "unmatched path omits http.route",
			target: "/wp-admin/setup-config.php",
			want: map[string]string{
				"http.request.method":       "GET",
				"url.scheme":                "http",
				"http.response.status_code": "404",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// A tagged post and an assets filesystem so every route under test
			// is registered.
			srv, collected := newMetricsServer(t,
				config.WithAssetsDir(testAssetsFS()).AsServerOption(),
			)

			get(srv, tt.target)

			points := durationPoints(t, collected(t))
			if len(points) != 1 {
				t.Fatalf("GET %s: got %d duration data points, want 1", tt.target, len(points))
			}
			got := attrs(points[0].Attributes)
			if len(got) != len(tt.want) {
				t.Errorf("GET %s: attributes %v, want %v", tt.target, got, tt.want)
			}
			for k, want := range tt.want {
				if got[k] != want {
					t.Errorf("GET %s: attribute %s = %q, want %q", tt.target, k, got[k], want)
				}
			}
		})
	}
}

// TestMetrics_ResponseBodySize verifies the body size histogram records the
// number of bytes actually written.
func TestMetrics_ResponseBodySize(t *testing.T) {
	t.Parallel()

	srv, collected := newMetricsServer(t)

	w := get(srv, "/")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /: status %d, want 200", w.Code)
	}
	want := int64(w.Body.Len())

	m := findMetric(t, collected(t), responseSizeMetric)
	hist, ok := m.Data.(metricdata.Histogram[int64])
	if !ok {
		t.Fatalf("%s: data is %T, want Histogram[int64]", responseSizeMetric, m.Data)
	}
	if len(hist.DataPoints) != 1 {
		t.Fatalf("%s: got %d data points, want 1", responseSizeMetric, len(hist.DataPoints))
	}
	if got := hist.DataPoints[0].Sum; got != want {
		t.Errorf("%s: recorded %d bytes, want %d", responseSizeMetric, got, want)
	}
}

// TestMetrics_ActiveRequestsBalance verifies the up/down counter returns to
// zero once requests have completed.
func TestMetrics_ActiveRequestsBalance(t *testing.T) {
	t.Parallel()

	srv, collected := newMetricsServer(t)

	for range 5 {
		get(srv, "/")
	}

	m := findMetric(t, collected(t), activeMetric)
	sum, ok := m.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("%s: data is %T, want Sum[int64]", activeMetric, m.Data)
	}
	for _, dp := range sum.DataPoints {
		if dp.Value != 0 {
			t.Errorf("%s: got %d in flight after all requests completed, want 0", activeMetric, dp.Value)
		}
	}
}

// TestMetrics_UnmatchedPathsBounded is the cardinality regression test: many
// distinct paths that match no route must collapse into a single time series.
// Labelling by r.URL.Path instead of the matched pattern would produce one
// series per request and eventually take a Prometheus instance down.
func TestMetrics_UnmatchedPathsBounded(t *testing.T) {
	t.Parallel()

	srv, collected := newMetricsServer(t)

	const requests = 500
	for i := range requests {
		target := fmt.Sprintf("/.scan-%d/%d/wp-admin", i, i*7919)
		if w := get(srv, target); w.Code != http.StatusNotFound {
			t.Fatalf("GET %s: status %d, want 404", target, w.Code)
		}
	}

	points := durationPoints(t, collected(t))
	if len(points) != 1 {
		t.Fatalf("%d unmatched requests produced %d time series, want 1", requests, len(points))
	}
	if points[0].Count != requests {
		t.Errorf("got count %d, want %d", points[0].Count, requests)
	}
	if _, ok := points[0].Attributes.Value("http.route"); ok {
		t.Error("unmatched requests carry an http.route attribute; they must share one series without it")
	}
}

// TestMetrics_UnknownMethodBounded verifies that arbitrary request methods
// collapse to the semconv "_OTHER" value rather than creating a series each.
func TestMetrics_UnknownMethodBounded(t *testing.T) {
	t.Parallel()

	srv, collected := newMetricsServer(t)

	for i := range 50 {
		req := httptest.NewRequest(fmt.Sprintf("FROB%d", i), "/", nil)
		srv.ServeHTTP(httptest.NewRecorder(), req)
	}

	points := durationPoints(t, collected(t))
	if len(points) != 1 {
		t.Fatalf("50 unknown methods produced %d time series, want 1", len(points))
	}
	if got := attrs(points[0].Attributes)["http.request.method"]; got != "_OTHER" {
		t.Errorf("http.request.method = %q, want %q", got, "_OTHER")
	}
}

// TestMetrics_HealthChecksNotRecorded pins the behaviour that /healthz/*
// requests are answered before the handler stack and therefore never recorded.
// Probe traffic would otherwise swamp real page hits.
func TestMetrics_HealthChecksNotRecorded(t *testing.T) {
	t.Parallel()

	srv, collected := newMetricsServer(t, config.WithHealthChecks())

	// Health checks defer content loading to Run, so /healthz/live answers 200
	// while the readiness probes answer 503. Either way they are intercepted in
	// ServeHTTP before the instrumented handler.
	wantStatus := map[string]int{
		"/healthz/live":    http.StatusOK,
		"/healthz/ready":   http.StatusServiceUnavailable,
		"/healthz/startup": http.StatusServiceUnavailable,
	}
	for path, want := range wantStatus {
		if w := get(srv, path); w.Code != want {
			t.Fatalf("GET %s: status %d, want %d", path, w.Code, want)
		}
	}

	if ms := collected(t); len(ms) != 0 {
		t.Fatalf("health probes recorded metrics %v, want none", metricNames(ms))
	}
}

// TestMetrics_DisabledByDefault verifies that a server created without
// config.WithMeterProvider records nothing, including via the global provider:
// instrumentation must be explicitly opted into.
func TestMetrics_DisabledByDefault(t *testing.T) {
	// Not parallel: this test installs a global meter provider.
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() {
		otel.SetMeterProvider(noop.NewMeterProvider())
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("provider.Shutdown: %v", err)
		}
	})
	otel.SetMeterProvider(provider)

	cfg := config.ServerConfig{
		Gen: []config.GeneratorOption{config.WithRawOutput()},
	}
	srv, err := server.New(slog.New(slog.DiscardHandler), createTestFS(t), cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	if srv.MeterProvider.Enabled() {
		t.Error("MeterProvider.Enabled() is true with no option supplied, want false")
	}
	if srv.MeterProvider.MeterProvider == nil {
		t.Error("MeterProvider is nil; the no-op provider should be the default")
	}

	for range 3 {
		if w := get(srv, "/"); w.Code != http.StatusOK {
			t.Fatalf("GET /: status %d, want 200", w.Code)
		}
	}

	if ms := collect(t, reader); len(ms) != 0 {
		t.Errorf("recorded %v without WithMeterProvider, want nothing", metricNames(ms))
	}
}

// TestMetrics_ExplicitNoopRecordsNothing verifies that passing the no-op
// provider explicitly is equivalent to omitting the option.
func TestMetrics_ExplicitNoopRecordsNothing(t *testing.T) {
	t.Parallel()

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{config.WithMeterProvider(noop.NewMeterProvider())},
		Gen:    []config.GeneratorOption{config.WithRawOutput()},
	}
	srv, err := server.New(slog.New(slog.DiscardHandler), createTestFS(t), cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	if srv.MeterProvider.Enabled() {
		t.Error("MeterProvider.Enabled() is true for the no-op provider, want false")
	}
	if w := get(srv, "/"); w.Code != http.StatusOK {
		t.Errorf("GET /: status %d, want 200", w.Code)
	}
}

// TestMetrics_PersistAcrossUpdates verifies the middleware is reinstalled when
// UpdatePosts swaps the handler.
func TestMetrics_PersistAcrossUpdates(t *testing.T) {
	t.Parallel()

	srv, collected := newMetricsServer(t)

	if err := srv.UpdatePosts(createTestFS(t), context.Background()); err != nil {
		t.Fatalf("UpdatePosts: %v", err)
	}

	get(srv, "/")

	if points := durationPoints(t, collected(t)); len(points) != 1 {
		t.Fatalf("after UpdatePosts: got %d duration data points, want 1", len(points))
	}
}

// TestMetrics_ConcurrentWithHandlerSwap hits the instruments from several
// goroutines while UpdatePosts hot-swaps the handler underneath, so the race
// detector can catch unsynchronised access in the middleware or the
// ResponseWriter wrapper.
func TestMetrics_ConcurrentWithHandlerSwap(t *testing.T) {
	t.Parallel()

	const (
		readers    = 8
		iterations = 100
	)

	srv, collected := newMetricsServer(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	for range readers {
		wg.Go(func() {
			for range iterations {
				if w := get(srv, "/posts/test-post"); w.Code != http.StatusOK {
					t.Errorf("reader: status %d, want 200", w.Code)
					return
				}
			}
		})
	}
	wg.Go(func() {
		for range iterations {
			if err := srv.UpdatePosts(metricsPostsFS(), ctx); err != nil {
				t.Errorf("UpdatePosts: %v", err)
				return
			}
		}
	})
	wg.Wait()

	points := durationPoints(t, collected(t))
	if len(points) != 1 {
		t.Fatalf("got %d duration data points, want 1", len(points))
	}
	if want := uint64(readers * iterations); points[0].Count != want {
		t.Errorf("got count %d, want %d", points[0].Count, want)
	}
}

// TestMetrics_AssetRangeRequest verifies that wrapping the ResponseWriter does
// not break the range-request handling http.FileServerFS provides for assets.
func TestMetrics_AssetRangeRequest(t *testing.T) {
	t.Parallel()

	srv, _ := newMetricsServer(t, config.WithAssetsDir(testAssetsFS()).AsServerOption())

	req := httptest.NewRequest(http.MethodGet, "/images/pipeline.png", nil)
	req.Header.Set("Range", "bytes=1-3")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusPartialContent {
		t.Fatalf("range request: status %d, want 206", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if want := pngBytes[1:4]; string(body) != string(want) {
		t.Errorf("range request body: got %q, want %q", body, want)
	}
	if got, want := w.Header().Get("Content-Range"), fmt.Sprintf("bytes 1-3/%d", len(pngBytes)); got != want {
		t.Errorf("Content-Range: got %q, want %q", got, want)
	}
}
