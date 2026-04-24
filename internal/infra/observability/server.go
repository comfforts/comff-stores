package observability

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/comfforts/logger"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const DEFAULT_METRICS_ADDR = ":9464"

type InitOptions struct {
	ServiceName   string
	MetricsAddr   string // e.g. ":9464" (Prom’s Prometheus exporter defaults to 9464)
	OTLPEndpoint  string // e.g. "otel-collector:4317" or "" to skip tracing exporter
	MetricsHandle string // e.g. "/metrics" (defaults to /metrics)
}

type metricsServer struct {
	srv *http.Server
	tp  *sdktrace.TracerProvider
	mp  *sdkmetric.MeterProvider
}

func NewMetricsServer(ctx context.Context, opt InitOptions) (*metricsServer, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	host, err := os.Hostname()
	if err != nil {
		l.Error("failed to get hostname", "error", err.Error())
		host = "unknown"
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(opt.ServiceName),
			semconv.ServiceInstanceIDKey.String(host),
		),
	)
	if err != nil {
		l.Error("failed to create resource for observability", "error", err.Error())
		return nil, err
	}

	// Create OTEL Prometheus exporter and tell it which registry to register into.
	// promExp, err := otelprom.New()
	reg := prom.NewRegistry()
	promExp, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		l.Error("failed to create Prometheus exporter", "error", err.Error())
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExp),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	// serve /metrics
	if opt.MetricsAddr == "" {
		opt.MetricsAddr = DEFAULT_METRICS_ADDR
	}

	if opt.MetricsHandle == "" {
		opt.MetricsHandle = "/metrics"
	}

	mux := http.NewServeMux()
	// mux.Handle(opt.MetricsHandle, promhttp.Handler())
	mux.Handle(opt.MetricsHandle, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	metricsServ := &http.Server{
		Addr:         opt.MetricsAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	var tp *sdktrace.TracerProvider
	if opt.OTLPEndpoint != "" {
		exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(opt.OTLPEndpoint), otlptracegrpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.TraceContext{})
	}

	go func() {
		if err := metricsServ.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				l.Error("metrics server failed to start", "error", err.Error())
			} else {
				l.Info("metrics server stopped")
			}
		}
	}()

	return &metricsServer{
		srv: metricsServ,
		tp:  tp,
		mp:  mp,
	}, nil
}

func (m *metricsServer) Shutdown(ctx context.Context) error {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	err = m.srv.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Error("Error shutting down metrics server", "error", err.Error())
		return err
	}
	if m.tp != nil {
		if tpErr := m.tp.Shutdown(ctx); tpErr != nil {
			l.Error("Error shutting down tracer provider", "error", tpErr.Error())
			err = errors.Join(err, tpErr)
		}
	}
	if m.mp != nil {
		if mpErr := m.mp.Shutdown(ctx); mpErr != nil {
			l.Error("Error shutting down meter provider", "error", mpErr.Error())
			err = errors.Join(err, mpErr)
		}
	}
	return err
}
