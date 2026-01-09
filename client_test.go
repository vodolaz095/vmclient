package vmclient

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestClientAgainstRealVM(tt *testing.T) {
	var metricName = fmt.Sprintf("something{job=%q,when=\"%v\",unit=%q}", "vmclient", time.Now().Unix(), "test")
	client, errC := New(tt.Context(), Config{Address: DefaultEndpoint, Insecure: true, Headers: map[string]string{"a": "b"}})
	if errC != nil {
		tt.Errorf("error creating client: %s", errC)
		return
	}

	tt.Run("ping", func(t *testing.T) {
		err := client.Ping(t.Context())
		if err != nil {
			t.Errorf("error pinging: %s", err)
		}
		return
	})

	tt.Run("push", func(t *testing.T) {
		set := metrics.NewSet()
		set.GetOrCreateGauge(metricName, func() float64 {
			return 10
		})
		err := client.Push(t.Context(), set)
		if err != nil {
			t.Errorf("error pushing: %s", err)
			return
		}
		t.Logf("Metrics are pushed")
	})

	tt.Run("instant ok", func(t *testing.T) {
		t.Logf("Sleeping for 5 seconds...")
		time.Sleep(5 * time.Second)
		instants, err := client.Instant(t.Context(), `something{job="vmclient",unit="test"}`, time.Now(), DefaultStep)
		if err != nil {
			t.Errorf("error sending instant query: %s", err)
			return
		}
		if len(instants) == 0 {
			t.Logf("nothing returned")
			return
		}
		for i := range instants {
			assert.Equal(t, "vmclient", instants[i].Labels["job"])
			assert.Equal(t, "test", instants[i].Labels["unit"])
			assert.Equal(t, "something", instants[i].Name())
			assert.Equal(t, float64(10), instants[i].Value)
			assert.Contains(t, instants[i].Labels, "when")
			t.Logf("Data received: %v - %s = %v", i, instants[i].String(), instants[i].Value)
		}
	})

	tt.Run("instant error", func(t *testing.T) {
		instants, err := client.Instant(t.Context(), `something{job="vmclient",unit="test}`, time.Now(), DefaultStep)
		assert.Error(t, err, "error should be thrown")
		assert.Empty(t, instants, "something is returned for wrong query")
		assert.ErrorIs(t, err, ErrQueryError, "wrong error returned")
		var properOne Err
		assert.ErrorAs(t, err, &properOne, "error is not type-casted")
		assert.Equal(t, http.StatusUnprocessableEntity, properOne.Code)
		assert.Contains(t, properOne.Message, "cannot find closing quote")
		t.Logf("error is %s", properOne.Error())
	})

	tt.Run("range ok", func(t *testing.T) {
		t.Logf("Sleeping for 5 seconds...")
		time.Sleep(5 * time.Second)
		ranges, err := client.Range(t.Context(), `something{job="vmclient",unit="test"}`, time.Now().Add(-time.Hour), time.Now(), DefaultStep)
		if err != nil {
			t.Errorf("error sending instant query: %s", err)
			return
		}
		if len(ranges) == 0 {
			t.Logf("nothing returned")
			return
		}
		for i := range ranges {
			assert.Equal(t, "vmclient", ranges[i].Labels["job"])
			assert.Equal(t, "test", ranges[i].Labels["unit"])
			assert.Equal(t, "something", ranges[i].Name())
			assert.Contains(t, ranges[i].Labels, "when")
			for j := range ranges[i].Values {
				t.Logf("For %s values are %v on %s", ranges[i].String(),
					ranges[i].Values[j].Value, ranges[i].Values[j].Timestamp.Format(time.Stamp))
			}
		}
	})

	tt.Run("range error", func(t *testing.T) {
		lines, err := client.Range(t.Context(), `something{job="vmclient",unit="test}`, time.Now().Add(-time.Minute), time.Now(), DefaultStep)
		assert.Error(t, err, "error should be thrown")
		assert.Empty(t, lines, "something is returned for wrong query")
		assert.ErrorIs(t, err, ErrQueryError, "wrong error returned")
		var properOne Err
		assert.ErrorAs(t, err, &properOne, "error is not type-casted")
		assert.Equal(t, http.StatusUnprocessableEntity, properOne.Code)
		assert.Contains(t, properOne.Message, "cannot find closing quote")
		t.Logf("error is %s", properOne.Error())
	})

	tt.Run("close", func(t *testing.T) {
		errClosing := client.Close(t.Context())
		if errClosing != nil {
			t.Errorf("error closing: %s", errClosing)
		}
	})
}

func TestClientAgainstExampleOrg(t *testing.T) {
	_, err := New(t.Context(), Config{Address: "http://example.org"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnexpectedResponse, "wrong error")
	var properOne Err
	assert.ErrorAs(t, err, &properOne, "wrong error")
	assert.Equal(t, http.StatusNotFound, properOne.Code, "wrong status code")
}

func TestAgainstHttpMock(tt *testing.T) {
	mockTransport := httpmock.NewMockTransport()
	mockTransport.RegisterResponder(http.MethodGet, DefaultEndpoint+"/-/healthy",
		httpmock.NewStringResponder(http.StatusOK, "200th status code is enough to trick vm client"))

	body := instantRawResponse{
		Status: "success",
		Data: instantRespData{Result: []instantResult{
			{
				Metric: map[string]string{
					"__name__": "something",
					"job":      "vmclient",
					"unit":     "test",
					"when":     strconv.FormatInt(time.Now().Add(-time.Second).Unix(), 10),
				},
				Group: 1,
				Value: []any{1734677495.161, "10"},
			},
			{
				Metric: map[string]string{
					"__name__": "something",
					"job":      "vmclient",
					"unit":     "test",
					"when":     strconv.FormatInt(time.Now().Unix(), 10),
				},
				Group: 1,
				Value: []any{1734677495.161, "10"},
			},
		}},
	}
	responder, err := httpmock.NewJsonResponder(http.StatusOK, body)
	if err != nil {
		tt.Fatal(err)
	}
	matcherFunc := func(req *http.Request) bool {
		if req.URL.Query().Get("query") == "" {
			return false
		}
		return `something{job="vmclient",unit="test"}` == req.URL.Query().Get("query")
	}
	matcher := httpmock.NewMatcher("validateSingleQuery", matcherFunc)
	singleQueryRegex, err := regexp.Compile(`\/prometheus\/api\/v1\/query`)
	if err != nil {
		tt.Fatal(err)
	}
	mockTransport.RegisterRegexpMatcherResponder(http.MethodGet, singleQueryRegex, matcher, responder)

	client, errC := New(tt.Context(), Config{
		Address:    DefaultEndpoint,
		HttpClient: &http.Client{Transport: mockTransport},
	})
	if errC != nil {
		tt.Errorf("error creating client: %s", errC)
		return
	}

	tt.Run("ping", func(t *testing.T) {
		errP := client.Ping(t.Context())
		if errP != nil {
			t.Errorf("error pinging: %s", errP)
		}
		return
	})

	tt.Run("instant ok", func(t *testing.T) {
		instants, errI := client.Instant(t.Context(), `something{job="vmclient",unit="test"}`, time.Now(), DefaultStep)
		if errI != nil {
			t.Errorf("error sending instant query: %s", errI)
			return
		}
		assert.Len(t, instants, 2)
		for i := range instants {
			assert.Equal(t, "vmclient", instants[i].Labels["job"])
			assert.Equal(t, "test", instants[i].Labels["unit"])
			assert.Equal(t, "something", instants[i].Name())
			assert.Equal(t, float64(10), instants[i].Value)
			assert.Contains(t, instants[i].Labels, "when")
			t.Logf("Data received: %v - %s = %v", i, instants[i].String(), instants[i].Value)
		}
	})
}
