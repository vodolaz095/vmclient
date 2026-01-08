package main

import (
	"context"
	"log"

	"github.com/VictoriaMetrics/metrics"
	"github.com/vodolaz095/vmclient"
)

func main() {
	ctx := context.TODO()
	client, err := vmclient.New(ctx, vmclient.Config{
		Address:     vmclient.DefaultEndpoint,
		ExtraLabels: `unit="test"`,
	})
	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}
	err = client.PushGauge(context.TODO(), `something{job="vmclient_example"}`, 10)
	if err != nil {
		log.Fatalf("error pushing gauge: %s", err)
	}
	err = client.PushCounter(context.TODO(), `something_cnt{job="vmclient_example"}`, 10)
	if err != nil {
		log.Fatalf("error pushing counter: %s", err)
	}
	err = client.Push(context.TODO(), metrics.GetDefaultSet())
	if err != nil {
		log.Fatalf("error pushing set: %s", err)
	}
	err = client.Close(ctx)
	if err != nil {
		log.Fatalf("error closing client: %s", err)
	}
	log.Printf("all data pushed, check %s for details", vmclient.DefaultEndpoint+"/vmui/?#/?g0.range_input=30m&g0.end_input=2026-01-08T17%3A46%3A23&g0.relative_time=last_30_minutes&g0.tab=0&g0.expr=something&g1.expr=something_cnt&g1.range_input=30m&g1.relative_time=last_30_minutes&g1.tab=0")
}
