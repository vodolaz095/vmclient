package vmclient

import (
	"context"
	"fmt"
	"io"
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
