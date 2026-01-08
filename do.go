package vmclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
	"go.opentelemetry.io/otel/trace"
)

type doParams struct {
	query string
	start time.Time
	end   time.Time
	when  time.Time
	step  time.Duration
}

func (c *Client) do(initialCtx context.Context, operation string, params doParams) (resp *http.Response, err error) {
	ctx, span := otel.GetTracerProvider().Tracer("vmclient").Start(initialCtx, operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBClientConnectionPoolName(c.endpoint),
			semconv.DBSystemNameKey.String("Victoria Metrics")),
	)
	defer span.End()
	var endpoint string
	var u *url.URL

	switch operation {
	case "ping":
		// https://github.com/VictoriaMetrics/VictoriaMetrics/issues/3539#issuecomment-1366469760
		endpoint, err = url.JoinPath(c.endpoint, "-", "healthy")
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, err
		}
	case "instant":
		u, err = url.Parse(c.endpoint)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, fmt.Errorf("error parsing endpoint: %s", err)
		}
		u.Path += "prometheus/api/v1/query"
		args := url.Values{}
		args.Set("query", params.query)
		args.Set("time", strconv.FormatInt(params.when.Unix(), 10))
		args.Set("step", params.step.String())
		deadline, present := ctx.Deadline()
		if present {
			args.Set("timeout", time.Until(deadline).String())
		}
		u.RawQuery = args.Encode()
		endpoint = u.String()
		span.SetAttributes(semconv.DBQueryText(params.query),
			attribute.String("end", params.when.Format(time.ANSIC)),
			attribute.String("step", params.step.String()),
		)
	case "range":
		u, err = url.Parse(c.endpoint)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, fmt.Errorf("error parsing endpoint: %s", err)
		}
		u.Path += "prometheus/api/v1/query_range"
		args := url.Values{}
		args.Set("query", params.query)
		args.Set("start", strconv.FormatInt(params.start.Unix(), 10))
		args.Set("end", strconv.FormatInt(params.end.Unix(), 10))
		args.Set("step", params.step.String())
		deadline, present := ctx.Deadline()
		if present {
			args.Set("timeout", time.Until(deadline).String())
		}
		u.RawQuery = args.Encode()
		endpoint = u.String()
		span.SetAttributes(semconv.DBQueryText(params.query),
			attribute.String("end", params.when.Format(time.ANSIC)),
			attribute.String("step", params.step.String()),
		)
	default:
		return nil, fmt.Errorf("unknown operation %s", operation)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	res, err := c.hclient.Do(req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	return res, nil
}
