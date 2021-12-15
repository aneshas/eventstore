package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/example/account"
)

func main() {
	estore, err := eventstore.New(
		"exampledb",
		eventstore.NewJSONEncoder(
			account.NewAccountOpenned{},
		),
	)
	checkErr(err)

	defer estore.Close()

	p := eventstore.NewProjector(estore)

	p.Add(
		NewConsoleOutputProjection(),
	)

	log.Fatal(p.Run(context.Background()))
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// NewConsoleOutputProjection consutructs an example projection that outputs
// new accounts to the console. It might as well be to any kind of
// database, disk, memory etc...
func NewConsoleOutputProjection() eventstore.Projection {
	return func(data eventstore.EventData) error {
		switch data.Event.(type) {
		case account.NewAccountOpenned:
			evt := data.Event.(account.NewAccountOpenned)
			fmt.Printf("Account: #%s | Holder: <%s>\n", evt.ID, evt.Holder)
		default:
			fmt.Println("not interested in this event")
		}

		return nil
	}
}
