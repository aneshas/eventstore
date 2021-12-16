package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

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
		NewJSONFileProjection("accounts.json"),
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

// NewJSONFileProjection makes use of flush after projection in order to
// periodically write accounts to a json file on disk
func NewJSONFileProjection(fname string) eventstore.Projection {
	var accounts []string

	return eventstore.FlushAfter(
		func(data eventstore.EventData) error {
			switch data.Event.(type) {
			case account.NewAccountOpenned:
				evt := data.Event.(account.NewAccountOpenned)
				accounts = append(accounts, fmt.Sprintf("Account: #%s | Holder: <%s>", evt.ID, evt.Holder))
			default:
				fmt.Println("not interested in this event")
			}

			return nil
		},
		func() error {
			if len(accounts) == 0 {
				return nil
			}

			data, err := json.Marshal(accounts)
			if err != nil {
				return err
			}

			err = os.WriteFile(fname, data, os.ModeAppend|os.ModePerm)
			if err != nil {
				return err
			}

			return nil
		},
		200*time.Millisecond,
	)
}
