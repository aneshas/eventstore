package echoambar

import (
	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/ambar"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

// Wrap returns a func wrapper around Ambar projection handler which adapts it to echo.HandlerFunc
func Wrap(a *ambar.Ambar) func(projection eventstore.Projection) echo.HandlerFunc {
	return func(projection eventstore.Projection) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()

			req, err := io.ReadAll(r.Body)
			if err != nil {
				return err
			}

			err = a.Project(r.Context(), projection, req)
			if err != nil {
				// response based on error

				return err
			}

			return c.JSONBlob(http.StatusOK, []byte(`{
  "result": {
    "success": {}
  }
}`))
		}
	}
}
