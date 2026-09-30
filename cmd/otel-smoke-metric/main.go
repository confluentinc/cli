// otel-smoke-metric sends a single pass/fail gauge metric via OTLP gRPC.
// Usage: otel-smoke-metric <0|1> where 0 = test failed, 1 = test passed.
// It always exits 0 so it never breaks the CI pipeline.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"
)

func main() {
	if err := run(); err != nil {
		log.Printf("otel-smoke-metric: %v", err)
	}
	// Always exit 0 — never fail the pipeline.
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: otel-smoke-metric <0|1>")
	}
	result, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid result value %q: %w", os.Args[1], err)
	}
	if result != 0 && result != 1 {
		return fmt.Errorf("invalid result value %q: must be 0 or 1", os.Args[1])
	}

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otlp.telemetry.confluent.cloud:4317"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint),
	}

	// Send bearer token as a gRPC header on every export request.
	token := os.Getenv("SEMAPHORE_OIDC_TOKEN")
	if token != "" {
		opts = append(opts, otlpmetricgrpc.WithHeaders(map[string]string{
			"authorization": "Bearer " + token,
		}))
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_INSECURE") == "true" {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	} else {
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("creating OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", "cli-smoke"),
		),
	)
	if err != nil {
		return fmt.Errorf("creating resource: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(time.Second),
		)),
	)

	meter := mp.Meter("smoke-tests")

	command := os.Getenv("SMOKE_COMMAND")
	if command == "" {
		command = "environment_list"
	}

	attrs := []attribute.KeyValue{
		attribute.String("test_suite", "smoke"),
		attribute.String("repo", "cli"),
		attribute.String("os", os.Getenv("SMOKE_OS")),
		attribute.String("arch", os.Getenv("SMOKE_ARCH")),
		attribute.String("command", command),
		attribute.String("branch", os.Getenv("SEMAPHORE_GIT_BRANCH")),
		attribute.String("job_id", os.Getenv("SEMAPHORE_JOB_ID")),
	}

	gauge, err := meter.Int64ObservableGauge("cli_smoke_test_result")
	if err != nil {
		return fmt.Errorf("creating gauge: %w", err)
	}

	_, err = meter.RegisterCallback(func(_ context.Context, o otelmetric.Observer) error {
		o.ObserveInt64(gauge, result, otelmetric.WithAttributes(attrs...))
		return nil
	}, gauge)
	if err != nil {
		return fmt.Errorf("registering callback: %w", err)
	}

	// Shutdown triggers collection and export.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := mp.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutting down meter provider: %w", err)
	}

	return nil
}
