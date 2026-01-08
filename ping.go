package vmclient

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Ping checks if database accepts connections
func (c *Client) Ping(ctx context.Context) (err error) {
	span := trace.SpanFromContext(ctx)
	resp, err := c.do(ctx, "ping", doParams{})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, ErrWrongStatus.Error())
		span.RecordError(ErrWrongStatus)
		return ErrWrongStatus
	}
	span.SetStatus(codes.Ok, "database responding")
	return nil
}
