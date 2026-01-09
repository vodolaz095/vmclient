package vmclient

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
	"go.opentelemetry.io/otel/trace"
)

// Ping checks if database accepts connections
func (c *Client) Ping(initialCtx context.Context) (err error) {
	ctx, span := otel.GetTracerProvider().Tracer("vmclient").Start(initialCtx, "ping",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBClientConnectionPoolName(c.endpoint),
			semconv.DBSystemNameKey.String("Victoria Metrics")),
	)
	defer span.End()

	resp, err := c.do(ctx, "ping", doParams{})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		retErr := Err{
			Err:     ErrUnexpectedResponse,
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("unexpected status code %s", resp.Status),
		}
		body, errReading := io.ReadAll(resp.Body)
		if errReading != nil {
			retErr.Err = errReading
			span.SetStatus(codes.Error, retErr.Error())
			span.RecordError(retErr)
			return retErr
		}
		retErr.Response = string(body)
		span.SetStatus(codes.Error, retErr.Error())
		span.RecordError(retErr)
		return retErr
	}
	span.SetStatus(codes.Ok, "database responding")
	return nil
}
