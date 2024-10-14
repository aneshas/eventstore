package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore-example/account"
)

func main() {
	eventStore, err := eventstore.New(
		eventstore.NewJSONEncoder(account.Events...),
		eventstore.WithSQLiteDB("example.db"),
	)
	checkErr(err)

	defer func() {
		_ = eventStore.Close()
	}()

	p := eventstore.NewProjector(eventStore)

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

// NewConsoleOutputProjection constructs an example projection that outputs
// new accounts to the console. It might as well be to any kind of
// database, disk, memory etc...
func NewConsoleOutputProjection() eventstore.Projection {
	return func(data eventstore.StoredEvent) error {
		switch data.Event.(type) {
		case account.NewAccountOpened:
			evt := data.Event.(account.NewAccountOpened)
			fmt.Printf("Account: #%s | Holder: <%s>\n", evt.AccountID, evt.Holder)

		default:
			fmt.Println("not interested in this event")
		}

		return nil
	}
}

// NewJSONFileProjection makes use of flush after projection in order to
// periodically write accounts to a json file on disk
func NewJSONFileProjection(fName string) eventstore.Projection {
	var accounts []string

	return eventstore.FlushAfter(
		func(data eventstore.StoredEvent) error {
			switch data.Event.(type) {
			case account.NewAccountOpened:
				evt := data.Event.(account.NewAccountOpened)
				accounts = append(accounts, fmt.Sprintf("Account: #%s Holder: %s", evt.AccountID, evt.Holder))
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

			err = os.WriteFile(fName, data, os.ModeAppend|os.ModePerm)
			if err != nil {
				return err
			}

			return nil
		},

		3*time.Second,
	)
}
