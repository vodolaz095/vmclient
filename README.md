vmclient
==========================

[![Go](https://github.com/vodolaz095/vmclient/actions/workflows/go.yml/badge.svg)](https://github.com/vodolaz095/vmclient/actions/workflows/go.yml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/vodolaz095/vmclient)](https://pkg.go.dev/github.com/vodolaz095/vmclient?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/vodolaz095/vmclient)](https://goreportcard.com/report/github.com/vodolaz095/vmclient)

Simple HTTP client for Victoria Metrics.

Pushing data
=======================

```go

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


```


Instant query
=======================
Endpoint https://docs.victoriametrics.com/victoriametrics/keyconcepts/#instant-query is used


```go

package main

import (
	"context"
	"log"
	"time"

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
	
	metrics, err := client.Instant(context.TODO(), "something", time.Now(), 5*time.Minute)
	if err != nil {
		log.Fatalf("error pushing set: %s", err)
	}
	for i := range metrics {
		log.Printf("Metric %s had value %v on %s",
			metrics[i].String(), metrics[i].Value, metrics[i].Timestamp.Format(time.DateTime))
	}

	err = client.Close(ctx)
	if err != nil {
		log.Fatalf("error closing client: %s", err)
	}
}


```


Range query
=======================
Endpoint https://docs.victoriametrics.com/victoriametrics/keyconcepts/#range-query is used

```go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

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

	lines, err := client.Range(context.TODO(), "something", time.Now().Add(-time.Hour), time.Now(), 5*time.Minute)
	if err != nil {
		log.Fatalf("error pushing set: %s", err)
	}
	for i := range lines {
		fmt.Printf("Found line for metric %s...\n", lines[i].String())
		for j := range lines[i].Values {
			fmt.Printf("%v) value %v on %s\n",
				j, lines[i].Values[j].Value, lines[i].Values[j].Timestamp.Format(time.DateTime))
		}
	}

	err = client.Close(ctx)
	if err != nil {
		log.Fatalf("error closing client: %s", err)
	}
}


```
