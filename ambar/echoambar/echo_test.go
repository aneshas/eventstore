package echoambar_test

import (
	"context"
	"fmt"
	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/ambar"
	"github.com/aneshas/eventstore/ambar/echoambar"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testPayload = `{ "payload": {}}`

type projector struct {
	wantErr error
	data    []byte
}

// Project projects ambar event to provided projection
func (p *projector) Project(_ context.Context, _ eventstore.Projection, data []byte) error {
	p.data = data

	if p.wantErr != nil {
		return p.wantErr
	}

	return nil
}

func TestShould_Project_Successfully(t *testing.T) {
	var p projector

	rec, err := project(t, &p, testPayload)

	assert.NoError(t, err)
	assert.Equal(t, testPayload, string(p.data))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, ambar.SuccessResp, rec.Body.String())
}

func TestShould_Project_With_KeepItGoing(t *testing.T) {
	var p projector

	p.wantErr = ambar.ErrKeepItGoing

	rec, err := project(t, &p, testPayload)

	assert.NoError(t, err)
	assert.Equal(t, testPayload, string(p.data))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, ambar.KeepGoingResp, rec.Body.String())
}

func TestShould_Project_With_Retry(t *testing.T) {
	var p projector

	p.wantErr = ambar.ErrRetry

	rec, err := project(t, &p, testPayload)

	assert.NoError(t, err)
	assert.Equal(t, testPayload, string(p.data))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, ambar.RetryResp, rec.Body.String())
}

func TestShould_Project_With_Retry_As_Fallback(t *testing.T) {
	var p projector

	p.wantErr = fmt.Errorf("some arbitrary error")

	rec, err := project(t, &p, testPayload)

	assert.NoError(t, err)
	assert.Equal(t, testPayload, string(p.data))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, ambar.RetryResp, rec.Body.String())
}

func project(t *testing.T, p *projector, payload string) (*httptest.ResponseRecorder, error) {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := echoambar.Wrap(p)(nil)

	err := h(c)

	return rec, err
}
