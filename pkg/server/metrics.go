// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/semconv/v1.43.0/httpconv"
)

// scopeName is the OpenTelemetry instrumentation scope reported for the
// server's metrics.
const scopeName = "github.com/harrydayexe/GoBlog/v2/pkg/server"

// metrics holds the OpenTelemetry instruments the server records for each
// request. The instruments follow the stable HTTP server semantic conventions
// so stock Prometheus rules and Grafana dashboards understand them without
// per-deployment queries.
type metrics struct {
	duration     httpconv.ServerRequestDuration
	active       httpconv.ServerActiveRequests
	responseSize httpconv.ServerResponseBodySize
}

// newMetrics creates the request instruments from mp.
//
// It is only called when a real meter provider was supplied; with the default
// no-op provider the server installs no instrumentation at all.
func newMetrics(mp metric.MeterProvider) (*metrics, error) {
	meter := mp.Meter(scopeName)

	duration, err := httpconv.NewServerRequestDuration(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s instrument: %w", duration.Name(), err)
	}
	active, err := httpconv.NewServerActiveRequests(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s instrument: %w", active.Name(), err)
	}
	responseSize, err := httpconv.NewServerResponseBodySize(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s instrument: %w", responseSize.Name(), err)
	}

	return &metrics{duration: duration, active: active, responseSize: responseSize}, nil
}

// middleware returns a [middleware.Middleware] that records the request
// instruments around next.
//
// It is installed as the outermost layer of the server's handler stack, so the
// recorded duration covers the whole response including user middleware.
// Requests to /healthz/* are answered before the stack runs and are therefore
// never recorded.
func (m *metrics) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		method := requestMethod(r.Method)
		scheme := urlScheme(r)

		m.active.Add(ctx, 1, method, scheme)
		defer m.active.Add(ctx, -1, method, scheme)

		recorder := &metricsResponseWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(recorder, r)
		elapsed := time.Since(start)

		// The three instruments share the same optional attributes, so they are
		// built once. Keys come from semconv; http.request.method and url.scheme
		// are added by the Record calls themselves.
		attrs := make([]attribute.KeyValue, 0, 3)
		attrs = append(attrs, semconv.HTTPResponseStatusCode(recorder.status))
		// r.Pattern is the pattern ServeMux matched, set in place during
		// routing, and is empty when nothing matched. Using it rather than
		// r.URL.Path bounds the label to the routes the blog actually
		// registers, so a bot scanning for /wp-admin and friends cannot create
		// unbounded time series; every unmatched request shares one series
		// with no http.route attribute.
		if r.Pattern != "" {
			attrs = append(attrs, semconv.HTTPRoute(r.Pattern))
		}
		// Semantic conventions ask for error.type to carry the status code as a
		// string when a server error ended the request and nothing more
		// specific is known.
		if recorder.status >= http.StatusInternalServerError {
			attrs = append(attrs, semconv.ErrorTypeKey.String(strconv.Itoa(recorder.status)))
		}

		m.duration.Record(ctx, elapsed.Seconds(), method, scheme, attrs...)
		m.responseSize.Record(ctx, recorder.written, method, scheme, attrs...)
	})
}

// requestMethod maps an HTTP method onto its http.request.method attribute
// value. Methods outside the registered set collapse to "_OTHER" as the
// semantic conventions require, which also keeps a client sending arbitrary
// method tokens from multiplying time series.
func requestMethod(method string) httpconv.RequestMethodAttr {
	switch method {
	case http.MethodGet:
		return httpconv.RequestMethodGet
	case http.MethodHead:
		return httpconv.RequestMethodHead
	case http.MethodPost:
		return httpconv.RequestMethodPost
	case http.MethodPut:
		return httpconv.RequestMethodPut
	case http.MethodPatch:
		return httpconv.RequestMethodPatch
	case http.MethodDelete:
		return httpconv.RequestMethodDelete
	case http.MethodConnect:
		return httpconv.RequestMethodConnect
	case http.MethodOptions:
		return httpconv.RequestMethodOptions
	case http.MethodTrace:
		return httpconv.RequestMethodTrace
	case "QUERY":
		return httpconv.RequestMethodQuery
	default:
		return httpconv.RequestMethodOther
	}
}

// urlScheme returns the url.scheme attribute value for a server request.
// Server requests carry no scheme in their URL, so it is derived from whether
// the connection was served over TLS.
func urlScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

// metricsResponseWriter wraps an [http.ResponseWriter] to capture the status
// code and body size the metrics middleware records.
//
// It implements Unwrap so [http.ResponseController] reaches the underlying
// writer, and forwards ReadFrom so the copy fast path net/http uses for static
// files under [http.FileServerFS] is preserved.
type metricsResponseWriter struct {
	http.ResponseWriter

	status      int  // response status code; 200 unless WriteHeader says otherwise
	wroteHeader bool // whether the status has been settled
	written     int64
}

func (w *metricsResponseWriter) WriteHeader(status int) {
	if !w.wroteHeader {
		w.wroteHeader = true
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *metricsResponseWriter) Write(b []byte) (int, error) {
	w.wroteHeader = true
	n, err := w.ResponseWriter.Write(b)
	w.written += int64(n)
	return n, err
}

// ReadFrom keeps the wrapped writer's [io.ReaderFrom] implementation reachable,
// which is what lets net/http copy file bodies efficiently instead of shuttling
// them through 32KiB Write calls.
func (w *metricsResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	w.wroteHeader = true

	rf, ok := w.ResponseWriter.(io.ReaderFrom)
	if !ok {
		// writeOnly hides this type's ReadFrom from io.Copy, which would
		// otherwise call back into it and recurse forever.
		n, err := io.Copy(writeOnly{w}, r)
		return n, err
	}

	n, err := rf.ReadFrom(r)
	w.written += n
	return n, err
}

// Flush forwards to the wrapped writer when it supports flushing, so handlers
// that assert [http.Flusher] directly rather than going through
// [http.ResponseController] still reach the real writer.
func (w *metricsResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap returns the wrapped writer so [http.ResponseController] can find the
// Hijack and deadline implementations it needs.
func (w *metricsResponseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// writeOnly exposes only Write, hiding any other interfaces its value satisfies.
type writeOnly struct{ io.Writer }
