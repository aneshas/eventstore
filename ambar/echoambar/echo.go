package echoambar

import (
	"errors"
	"github.com/aneshas/eventstore/ambar"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

var _ Projector = (*ambar.Ambar)(nil)

// Projector is an interface for projecting events
type Projector interface {
	Project(r *http.Request, projection ambar.Projection, data []byte) error
}

// Wrap returns a func wrapper around Ambar projection handler which adapts it to echo.HandlerFunc
func Wrap(a Projector) func(projection ambar.Projection) echo.HandlerFunc {
	return func(projection ambar.Projection) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()

			req, err := io.ReadAll(r.Body)
			if err != nil {
				return err
			}

			err = a.Project(r, projection, req)
			if err != nil {
				if errors.Is(err, ambar.ErrNoRetry) {
					return c.JSONBlob(http.StatusOK, []byte(ambar.SuccessResp))
				}

				if errors.Is(err, ambar.ErrKeepItGoing) {
					return c.JSONBlob(http.StatusOK, []byte(ambar.KeepGoingResp))
				}

				return c.JSONBlob(http.StatusOK, []byte(ambar.RetryResp))
			}

			return c.JSONBlob(http.StatusOK, []byte(ambar.SuccessResp))
		}
	}
}
