package main

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	otelmetric "go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

// initProvider sets up the OpenTelemetry MeterProvider,TraceProvider and exports metrics.
func initProvider() (*metric.MeterProvider, *trace.TracerProvider, otelmetric.Int64Counter, otelmetric.Float64Histogram, error) {

	//  Configure Resource
	res, err := newResource()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}
	// Configure TraceProvider
	tracerProvider, err := newTraceProvider(res)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create meter provider: %w", err)
	}
	otel.SetTracerProvider(tracerProvider)

	//  Configure MeterProvider
	meterProvider, err := newMeterProvider(res)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create meter provider: %w", err)
	}

	otel.SetMeterProvider(meterProvider)
	meter := otel.Meter("chat-metrics")

	// Initialize counter metrics
	requestCounter, err := meter.Int64Counter(
		"api.counter",
		otelmetric.WithDescription("Number of API calls."),
		otelmetric.WithUnit("{call}"),
	)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create request counter: %w", err)
	}

	// Initialize duration metrics
	requestDuration, err := meter.Float64Histogram(
		"http.server.response_time",
		otelmetric.WithDescription("Duration of HTTP server requests"),
		otelmetric.WithUnit("ms"),
	)

	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create request duration histogram: %w", err)
	}

	return meterProvider, tracerProvider, requestCounter, requestDuration, nil
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("Clippy-chat"),
			semconv.ServiceVersion("1.0.1"),
		),
	)
}

func newMeterProvider(res *resource.Resource) (*metric.MeterProvider, error) {
	// Configure exporter
	metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(1*time.Minute))),
	)
	return meterProvider, nil
}

// TraceProvider
func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	// Configure exporter, same as metrics one
	traceExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(5*time.Second)),
	)
	return tracerProvider, nil
}
