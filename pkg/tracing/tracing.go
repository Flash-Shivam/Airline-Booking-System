package tracing

import (
	"context"
	"log"
	"time"

	"airline-booking-system/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracer configures a global tracer provider. It returns a shutdown
// function that should be deferred from main.
func InitTracer(ctx context.Context, cfg *config.TracingConfig) (func(context.Context) error, error) {
	if cfg == nil || !cfg.Enabled {
		// Tracing disabled; nothing to do.
		return func(context.Context) error { return nil }, nil
	}

	// Configure OTLP HTTP exporter (compatible with most OTel collectors and Jaeger via collector).
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(cfg.Endpoint),
		otlptracehttp.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(cfg.SamplerRatio),
		)),
	)

	otel.SetTracerProvider(tp)

	log.Printf("Tracing enabled: service=%s endpoint=%s env=%s", cfg.ServiceName, cfg.Endpoint, cfg.Environment)

	return tp.Shutdown, nil
}


