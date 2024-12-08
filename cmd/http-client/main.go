package main

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
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	ctx := context.Background()
	cleanup, err := otelDo(ctx)

	client := http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
				return otelhttptrace.NewClientTrace(ctx)
			}),
		),
	}

	resp, err := client.Get("https://scrapbox.io/tkdn")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		cleanup()
		resp.Body.Close()
	}()

	fmt.Println(resp.Status)
}

func otelDo(ctx context.Context) (func(), error) {
	attrs := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("example.httpclient"),
	)
	resc, err := resource.New(ctx,
		resource.WithAttributes(attrs.Attributes()...),
		resource.WithTelemetrySDK(),
	)

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(resc),
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
	return cleanup, nil
}
