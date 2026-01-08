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
