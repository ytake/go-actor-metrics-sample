package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type NrOpenTelemetry struct {
	endpoint  string
	attribute []attribute.KeyValue
	apiKey    string
}

// NewNrOpenTelemetry returns OpenTelemetry for newrelic
func NewNrOpenTelemetry(endpoint, serviceName, apiKey string) *NrOpenTelemetry {
	return &NrOpenTelemetry{
		endpoint: endpoint,
		attribute: []attribute.KeyValue{
			attribute.String("service.name", serviceName),
		},
		apiKey: apiKey,
	}
}

// Exporter returns metric.MeterProvider
func (o *NrOpenTelemetry) Exporter(ctx context.Context) (*metric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		otlpmetrichttp.WithEndpoint(o.endpoint),
		otlpmetrichttp.WithHeaders(map[string]string{
			"api-key": o.apiKey,
		}),
	)
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx, resource.WithAttributes(o.attribute...))
	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(1*time.Second))),
		metric.WithResource(res))
	otel.SetMeterProvider(provider)
	return provider, nil
}
