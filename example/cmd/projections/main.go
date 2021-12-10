package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/example/account"
)

func main() {
	estore, err := eventstore.New(
		"exampledb",
		eventstore.NewJsonEncoder(
			account.NewAccountOpenned{},
		),
	)
	checkErr(err)

	defer estore.Close()

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	sub, err := estore.ReadAll(ctx)
	checkErr(err)

	runConsoleOutputProjection(sub)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// An example projection that outputs new accounts to the console
// it might as well be any kind of database, disk, memory etc...
func runConsoleOutputProjection(sub eventstore.Subscription) {
	for {
		select {
		case data := <-sub.EventData:
			handle(data)

		case err := <-sub.Err:
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				log.Fatal(err)
			}
		}
	}
}

func handle(data eventstore.EventData) {
	switch data.Event.(type) {
	case account.NewAccountOpenned:
		evt := data.Event.(account.NewAccountOpenned)
		fmt.Printf("Account: #%s | Holder: <%s>\n", evt.ID, evt.Holder)
	default:
		fmt.Println("not interested in this event")
	}
}
