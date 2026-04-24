package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Metrics interface {
	AddInflightRequest(ctx context.Context, method string, delta int64)
	IncRequest(ctx context.Context, method, status string)
	ObserveRequestDuration(ctx context.Context, method, status string, duration time.Duration)
}

type metrics struct {
	scope            metric.Meter
	inflightRequests metric.Int64UpDownCounter
	requestCounter   metric.Int64Counter
	requestDuration  metric.Float64Histogram
}

func NewMetrics() (*metrics, error) {
	meter := otel.Meter("geo-grpc")
	inflight, err := meter.Int64UpDownCounter("geo_inflight_requests")
	if err != nil {
		return nil, err
	}
	reqs, err := meter.Int64Counter("geo_requests_total")
	if err != nil {
		return nil, err
	}
	reqDuration, err := meter.Float64Histogram("geo_request_duration_seconds")
	if err != nil {
		return nil, err
	}
	return &metrics{
		scope:            meter,
		inflightRequests: inflight,
		requestCounter:   reqs,
		requestDuration:  reqDuration,
	}, nil
}

func (m *metrics) AddInflightRequest(ctx context.Context, method string, delta int64) {
	m.inflightRequests.Add(ctx, delta,
		metric.WithAttributes(
			attribute.String("rpc.method", method),
		),
	)
}

func (m *metrics) IncRequest(ctx context.Context, method, status string) {
	m.requestCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("rpc.method", method),
			attribute.String("rpc.status", status),
		),
	)
}

func (m *metrics) ObserveRequestDuration(ctx context.Context, method, status string, duration time.Duration) {
	m.requestDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(
			attribute.String("rpc.method", method),
			attribute.String("rpc.status", status),
		),
	)
}
