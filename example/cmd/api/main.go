package main

import (
	"errors"
	"fmt"
	"github.com/aneshas/eventstore/aggregate"
	"github.com/labstack/echo/v4"
	"log"
	"math/rand"
	"net/http"
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

	e := echo.New()

	e.Use(mw())

	e.GET("/accounts/open", NewOpenAccountHandlerFunc(eventStore))
	e.GET("/accounts/:id/deposit", NewDepositToAccountHandlerFunc(eventStore))

	log.Fatal(e.Start(":8080"))
}

func mw() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := aggregate.CtxWithMeta(c.Request().Context(), map[string]string{
				"ip":  c.Request().RemoteAddr,
				"app": "example-app",
			})

			ctx = aggregate.CtxWithCausationID(ctx, "some-causation-event-id")
			ctx = aggregate.CtxWithCorrelationID(ctx, "some-correlation-event-id")

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

type newAccountResp struct {
	AccountID string `json:"account_id"`
}

// NewOpenAccountHandlerFunc creates new account opening endpoint example
func NewOpenAccountHandlerFunc(eventStore *eventstore.EventStore) echo.HandlerFunc {
	store := aggregate.NewStore[*account.Account](eventStore)

	return func(c echo.Context) error {
		id := fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Int())

		acc, err := account.New(account.ID(id), "John Doe")
		if err != nil {
			return err
		}

		err = store.Save(c.Request().Context(), acc)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, newAccountResp{AccountID: acc.ID()})
	}
}

type depositResp struct {
	AccountID  string `json:"account_id"`
	NewBalance int    `json:"new_balance"`
}

// NewDepositToAccountHandlerFunc creates new deposit to account endpoint example
func NewDepositToAccountHandlerFunc(eventStore *eventstore.EventStore) echo.HandlerFunc {
	store := aggregate.NewStore[*account.Account](eventStore)

	return func(c echo.Context) error {
		var (
			acc account.Account
			ctx = c.Request().Context()
		)

		err := store.FindByID(ctx, c.Param("id"), &acc)
		if err != nil {
			if errors.Is(err, aggregate.ErrAggregateNotFound) {
				return c.String(http.StatusNotFound, "Account not found")
			}

			return err
		}

		amount := 100

		acc.Deposit(amount)

		err = store.Save(ctx, &acc)
		if err != nil {
			return err
		}

		return c.JSON(
			http.StatusOK,
			depositResp{
				AccountID:  acc.ID(),
				NewBalance: acc.Balance,
			},
		)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
