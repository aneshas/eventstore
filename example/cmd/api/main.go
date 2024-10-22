package main

import (
	"errors"
	"flag"
	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore-example/account"
	"github.com/aneshas/eventstore/aggregate"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"os"
	"strconv"
)

var pg = flag.Bool("pg", false, "Run with postgres db (set env DSN to pg connection string)")

func main() {
	flag.Parse()

	db := eventstore.WithSQLiteDB("example.db")

	if *pg {
		db = eventstore.WithPostgresDB(os.Getenv("DSN"))
	}

	eventStore, err := eventstore.New(
		eventstore.NewJSONEncoder(account.Events...),
		db,
	)
	checkErr(err)

	defer func() {
		_ = eventStore.Close()
	}()

	e := echo.New()

	e.Use(mw())

	e.GET("/accounts/open", NewOpenAccountHandlerFunc(eventStore))
	e.GET("/accounts/:id/deposit/:amount", NewDepositToAccountHandlerFunc(eventStore))
	e.GET("/accounts/:id/withdraw/:amount", NewWithdrawFromAccountHandlerFunc(eventStore))

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if errors.Is(err, aggregate.ErrAggregateNotFound) {
			_ = c.String(http.StatusNotFound, "Account not found")
		}
	}

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
		acc, err := account.New(account.NewID(), "John Doe")
		if err != nil {
			return err
		}

		err = store.Save(c.Request().Context(), acc)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, newAccountResp{AccountID: acc.StringID()})
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

		err := store.ByID(ctx, c.Param("id"), &acc)
		if err != nil {
			return err
		}

		amount, _ := strconv.Atoi(c.Param("amount"))

		acc.Deposit(amount)

		err = store.Save(ctx, &acc)
		if err != nil {
			return err
		}

		return c.JSON(
			http.StatusOK,
			depositResp{
				AccountID:  acc.StringID(),
				NewBalance: acc.Balance,
			},
		)
	}
}

type withdrawResp struct {
	AccountID  string `json:"account_id"`
	NewBalance int    `json:"new_balance"`
}

// NewWithdrawFromAccountHandlerFunc creates new withdraw from account endpoint example
func NewWithdrawFromAccountHandlerFunc(eventStore *eventstore.EventStore) echo.HandlerFunc {
	store := aggregate.NewStore[*account.Account](eventStore)

	return func(c echo.Context) error {
		var (
			acc account.Account
			ctx = c.Request().Context()
		)

		err := store.ByID(ctx, c.Param("id"), &acc)
		if err != nil {
			return err
		}

		amount, _ := strconv.Atoi(c.Param("amount"))

		err = acc.Withdraw(amount)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		err = store.Save(ctx, &acc)
		if err != nil {
			return err
		}

		return c.JSON(
			http.StatusOK,
			withdrawResp{
				AccountID:  acc.StringID(),
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
