package example

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/aneshas/eventstore-example/account"
)

// NewOpenAccountHandlerFunc creates new account openning endpoint example
func NewOpenAccountHandlerFunc(store *AccountStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Int())

		acc, err := account.New(id, "John Doe")
		if err != nil {
			fmt.Fprintf(rw, "error: %v", err)
		}

		err = store.Save(r.Context(), acc)
		if err != nil {
			fmt.Fprintf(rw, "error: %v", err)
		}

		fmt.Fprintf(rw, "account created: %s", acc.ID)
	}
}
