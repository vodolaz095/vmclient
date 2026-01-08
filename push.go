package vmclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/VictoriaMetrics/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
	"go.opentelemetry.io/otel/trace"
)

// Push sends metrics set
func (c *Client) Push(initialCtx context.Context, set *metrics.Set) error {
	ctx, span := otel.GetTracerProvider().Tracer("vmclient").Start(initialCtx, "push",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBClientConnectionPoolName(c.endpoint),
			semconv.DBSystemNameKey.String("Victoria Metrics"),
			attribute.StringSlice("metric.names", set.ListMetricNames()),
		),
	)
	defer span.End()

	endpoint, err := url.JoinPath(c.endpoint, DefaultPushEndpoint)
	if err != nil {
		return err
	}
	var i int
	headers := make([]string, len(c.headers))
	for k := range c.headers {
		headers[i] = k + ": " + c.headers[k]
		i++
	}
	return set.PushMetrics(ctx, endpoint, &metrics.PushOptions{
		ExtraLabels: c.extraLabels,
		Headers:     headers,
		Method:      http.MethodPost,
	})
}

// PushGauge pushes metrics gauge
func (c *Client) PushGauge(ctx context.Context, name string, value float64) error {
	set := metrics.NewSet()
	set.GetOrCreateGauge(name, func() float64 {
		return value
	})
	return c.Push(ctx, set)
}

// PushCounter pushes counter
func (c *Client) PushCounter(ctx context.Context, name string, value uint64) error {
	set := metrics.NewSet()
	set.GetOrCreateCounter(name).Set(value)
	return c.Push(ctx, set)
}
