package http

import (
	"fmt"
	"github.com/gcjensen/splend"
	"github.com/gcjensen/splend/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSettleOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(randomUser(), dbh)
	user.AddOutgoing(randomOutgoing())
	outgoings, _ := user.GetOutgoings()
	outgoing := outgoings[0]

	router := httprouter.New()
	router.POST(
		"/outgoing/settle/:outgoingID/:shouldSettle",
		SettleOutgoing(dbh),
	)

	req, _ := http.NewRequest(
		"POST", fmt.Sprintf("/outgoing/settle/%d/1", *outgoing.ID), nil,
	)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedResponse := `{"message":"Outgoing settled!"}`
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponse)
	}

	// Test un-settling

	req, _ = http.NewRequest(
		"POST", fmt.Sprintf("/outgoing/settle/%d/0", *outgoing.ID), nil,
	)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	expectedResponse = `{"message":"Outgoing un-settled!"}`
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponse)
	}
}
