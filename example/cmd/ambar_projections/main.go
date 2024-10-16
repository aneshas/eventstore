package main

import (
	"fmt"
	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore-example/account"
	"github.com/aneshas/eventstore/ambar"
	"github.com/aneshas/eventstore/ambar/echoambar"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
)

func main() {
	e := echo.New()

	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == "user" && password == "pass" {
			return true, nil
		}

		return false, nil
	}))

	hf := echoambar.Wrap(
		ambar.New(eventstore.NewJSONEncoder(eventSubscriptions...)),
	)

	e.POST("/projections/accounts/v1", hf(NewConsoleOutputProjection()))

	log.Fatal(e.Start(":8181"))
}

var eventSubscriptions = []any{
	account.NewAccountOpened{},
	account.DepositMade{},
	account.WithdrawalMade{},
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

		case account.DepositMade:
			evt := data.Event.(account.DepositMade)
			fmt.Printf("Deposited the amount of %d EUR\n", evt.Amount)

		default:
			fmt.Println("not interested in this event")
		}

		return nil
	}
}
