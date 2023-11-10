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

type OpenTelemetry struct {
	endpoint  string
	attribute []attribute.KeyValue
}

// NewOpenTelemetry returns OpenTelemetry
func NewOpenTelemetry(endpoint, serviceName string) *OpenTelemetry {
	return &OpenTelemetry{
		endpoint: endpoint,
		attribute: []attribute.KeyValue{
			attribute.String("service.name", serviceName),
		},
	}
}

// Exporter returns metric.MeterProvider
func (o *OpenTelemetry) Exporter(ctx context.Context) (*metric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(o.endpoint),
		otlpmetrichttp.WithInsecure())
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
