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
