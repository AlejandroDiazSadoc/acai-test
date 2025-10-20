package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(requestCounter metric.Int64Counter, requestDuration metric.Float64Histogram) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrappedWriter := newResponseWriterWrapper(w)
			next.ServeHTTP(wrappedWriter, r) // Call the actual handler

			// --- Metric Capture Logic ---

			duration := time.Since(start)

			var route string
			if currentRoute := mux.CurrentRoute(r); currentRoute != nil {
				route, _ = currentRoute.GetPathTemplate()
			} else {
				route = "unmatched"
			}

			statusCode := wrappedWriter.statusCode

			slog.Info("Request finished, attempting metric export.", "route", route, "status", statusCode)

			commonAttributes := []attribute.KeyValue{
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.HTTPRoute(route),
				semconv.HTTPResponseStatusCode(statusCode),
			}

			// Record Metrics (This is where the variables are used, satisfying the compiler)
			requestCounter.Add(r.Context(), 1, metric.WithAttributes(commonAttributes...))
			requestDuration.Record(r.Context(), float64(duration.Milliseconds()), metric.WithAttributes(commonAttributes...))
		})
	}
}
