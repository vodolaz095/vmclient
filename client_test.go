package vmclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/stretchr/testify/assert"
)

func TestClient(tt *testing.T) {
	var metricName = fmt.Sprintf("something{job=%q,when=\"%v\",unit=%q}", "vmclient", time.Now().Unix(), "test")
	client, errC := New(tt.Context(), Config{Address: DefaultEndpoint})
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

	tt.Run("instant", func(t *testing.T) {
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

	tt.Run("range", func(t *testing.T) {
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
					ranges[i].Values[j].Value,
					ranges[i].Values[j].Timestamp.Format(time.Stamp),
				)
			}
		}
	})
}
