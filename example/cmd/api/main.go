package main

import (
	"log"
	"net/http"

	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/example"
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

	store := example.NewAccountStore(estore)

	http.Handle(
		"/accounts/open",
		example.NewOpenAccountHandlerFunc(store),
	)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
