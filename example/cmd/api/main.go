package main

import (
	"gorm.io/driver/sqlite"
	"log"
	"net/http"

	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore-example"
	"github.com/aneshas/eventstore-example/account"
)

func main() {
	estore, err := eventstore.New(
		sqlite.Open("exampledb"),
		eventstore.NewJSONEncoder(
			account.NewAccountOpened{},
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
