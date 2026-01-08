package vmclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type instantResult struct {
	Metric map[string]string `json:"metric"`
	Group  int               `json:"group"`
	Value  []any             `json:"value"`
}

func (m *instantResult) convert() (output Instant, err error) {
	output = Instant{}
	output.Labels = m.Metric
	if len(m.Value) > 0 {
		rawTimeStamp, ok := m.Value[0].(float64)
		if !ok {
			return output, fmt.Errorf("error parsing %s as float64", m.Value[0])
		}
		output.Timestamp = time.UnixMilli(int64(1000 * rawTimeStamp))
	}
	if len(m.Value) > 1 {
		stringified, ok := m.Value[1].(string)
		if !ok {
			return output, fmt.Errorf("error typecasting %v to string", m.Value[1])
		}
		rawValueParsed, errParsing := strconv.ParseFloat(stringified, 64)
		if errParsing != nil {
			return output, fmt.Errorf("error parsing value %s: %w", stringified, errParsing)
		}
		output.Value = rawValueParsed
	}
	return output, nil
}

type instantRespData struct {
	Result []instantResult `json:"result"`
}

type instantRawResponse struct {
	Status string          `json:"status"`
	Data   instantRespData `json:"data"`
}

// Instant makes instant query described here
// https://docs.victoriametrics.com/victoriametrics/keyconcepts/#instant-query
func (c *Client) Instant(ctx context.Context, query string, when time.Time, step time.Duration) (data []Instant, err error) {
	var output Instant
	span := trace.SpanFromContext(ctx)
	resp, err := c.do(ctx, "instant", doParams{query: query, when: when, step: step})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, ErrWrongStatus.Error())
		span.RecordError(ErrWrongStatus)
		return nil, ErrWrongStatus
	}
	span.AddEvent("request performed")
	var raw instantRawResponse
	err = json.NewDecoder(resp.Body).Decode(&raw)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("body parsed")
	if raw.Status != "success" {
		err = fmt.Errorf("wrong status: %s", raw.Status)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	data = make([]Instant, len(raw.Data.Result))
	for i := range raw.Data.Result {
		output, err = raw.Data.Result[i].convert()
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, err
		}
		data[i] = output
	}
	span.SetStatus(codes.Ok, "data received")
	return data, nil
}
