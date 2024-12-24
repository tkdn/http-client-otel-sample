package instrument

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const serviceName = "example"

func Setup(ctx context.Context) (*sdktrace.TracerProvider, func(), error) {
	attrs := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)
	resc, err := resource.New(ctx,
		resource.WithAttributes(attrs.Attributes()...),
		resource.WithTelemetrySDK(),
	)

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resc),
	)
	otel.SetTracerProvider(provider)
	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := provider.ForceFlush(ctx); err != nil {
			fmt.Println(err.Error())
		}
		defer cancel()
		if err := provider.Shutdown(shutdownCtx); err != nil {
			fmt.Println(err.Error())
		}
	}
	return provider, cleanup, nil
}

func NewHTTPTransport(tp *sdktrace.TracerProvider) http.RoundTripper {
	return otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithClientTrace(HTTPClientTracer),
	)
}

func HTTPClientTracer(ctx context.Context) *httptrace.ClientTrace {
	return otelhttptrace.NewClientTrace(ctx)
}

func HTTPConnTracer(tp *sdktrace.TracerProvider) func(context.Context) *httptrace.ClientTrace {
	return func(ctx context.Context) *httptrace.ClientTrace {
		var span trace.Span
		return &httptrace.ClientTrace{
			GetConn: func(_ string) {
				_, span = tp.Tracer("めっちゃ大事なトレース").Start(ctx, "めっちゃ大事なスパン")
			},
			GotConn: func(_ httptrace.GotConnInfo) {
				if span != nil {
					span.End()
				}
			},
		}
	}
}
