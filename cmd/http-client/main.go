package main

import (
	"context"
	"fmt"

	"github.com/tkdn/http-client-otel-sample/gateway"
	"github.com/tkdn/http-client-otel-sample/instrument"
)

func main() {
	ctx := context.Background()

	tp, cleanup, err := instrument.Setup(ctx)
	tr := instrument.NewHTTPTransport(tp)

	client := gateway.NewHTTPClient(tr)
	resp, err := client.Get(ctx, "https://scrapbox.io/tkdn")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer cleanup()

	fmt.Println(resp.Status)
}
