package vmclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	// ErrUnexpectedResponse happens, when Victoria Metrics gives unexpected response
	ErrUnexpectedResponse = errors.New("unexpected response")
	// ErrQueryError happens, when Victoria Metrics cannot process query
	ErrQueryError = errors.New("query error")
)

// Err is custom error
type Err struct {
	Code     int
	Message  string
	Response string
	Err      error
}

func (ei Err) Error() string {
	return fmt.Sprintf("%v - %s", ei.Code, ei.Message)
}

func (ei Err) Is(target error) bool {
	if ei.Err == target {
		return true
	}
	return ei == target
}

func (ei Err) Unwrap() error {
	return ei.Err
}

func (ei Err) As(target any) bool {
	return errors.As(ei, &target)
}

type errorResponse struct {
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Message   string `json:"error"`
}

func handleErrorResponse(resp *http.Response, span trace.Span) error {
	if resp.StatusCode == http.StatusOK {
		span.AddEvent("response code is correct")
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return Err{
			Code:     resp.StatusCode,
			Message:  "error reading body",
			Response: "",
			Err:      err,
		}
	}
	if resp.StatusCode == http.StatusUnprocessableEntity {
		var eResp errorResponse
		err = json.Unmarshal(raw, &eResp)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return Err{
				Code:     http.StatusUnprocessableEntity,
				Message:  fmt.Sprintf("error parsing response: %s", err),
				Response: string(raw),
				Err:      ErrUnexpectedResponse,
			}
		}
		span.SetStatus(codes.Error, eResp.Message)
		err = Err{
			Code:     http.StatusUnprocessableEntity,
			Message:  eResp.Message,
			Response: string(raw),
			Err:      ErrQueryError,
		}
		span.RecordError(err)
		return err
	}

	err = Err{
		Code:     resp.StatusCode,
		Err:      ErrUnexpectedResponse,
		Message:  "unexpected response",
		Response: string(raw),
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}
