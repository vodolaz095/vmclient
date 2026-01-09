package vmclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type rangeResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]any           `json:"values"`
}

func parseRangeValue(input []any) (Result, error) {
	var ret Result
	if len(input) != 2 {
		return ret, fmt.Errorf("exactly two parameters are expected, instead of %v", input)
	}

	rawTimeStamp, ok := input[0].(float64)
	if !ok {
		return ret, fmt.Errorf("error parsing %s as float64", input[0])
	}
	ret.Timestamp = time.UnixMilli(int64(1000 * rawTimeStamp))

	stringified, ok := input[1].(string)
	if !ok {
		return ret, fmt.Errorf("error typecasting %v to string", input[1])
	}
	rawValueParsed, errParsing := strconv.ParseFloat(stringified, 64)
	if errParsing != nil {
		return ret, fmt.Errorf("error parsing value %s: %w", stringified, errParsing)
	}
	ret.Value = rawValueParsed
	return ret, nil
}

type rangeRespData struct {
	Result []rangeResult `json:"result"`
}

type rangeRawResponse struct {
	Status string        `json:"status"`
	Data   rangeRespData `json:"data"`
}

// Range makes range query as described here https://docs.victoriametrics.com/victoriametrics/keyconcepts/#range-query
func (c *Client) Range(ctx context.Context, query string, start, end time.Time, step time.Duration) (data []Range, err error) {
	var result Result
	span := trace.SpanFromContext(ctx)
	resp, err := c.do(ctx, "range", doParams{query: query, start: start, end: end, step: step})
	if err != nil {
		return nil, err
	}
	err = handleErrorResponse(resp, span)
	if err != nil {
		return nil, err
	}
	span.AddEvent("request performed")
	var raw rangeRawResponse
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
	data = make([]Range, len(raw.Data.Result))
	for i := range raw.Data.Result {
		values := make([]Result, len(raw.Data.Result[i].Values))
		for j := range raw.Data.Result[i].Values {
			result, err = parseRangeValue(raw.Data.Result[i].Values[j])
			if err != nil {
				return nil, err
			}
			values[j] = result
		}
		data[i] = Range{
			Labels: raw.Data.Result[i].Metric,
			Values: values,
		}
	}
	span.SetStatus(codes.Ok, "data received")
	return data, nil
}
