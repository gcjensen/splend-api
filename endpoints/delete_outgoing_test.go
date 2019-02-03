package endpoints

import (
	"fmt"
	"github.com/gcjensen/splend-api/config"
	"github.com/gcjensen/splend-api/user"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := user.New(randomUser(), dbh)
	user.AddOutgoing(randomOutgoing())
	outgoings, _ := user.GetOutgoings()

	router := httprouter.New()
	router.POST("/outgoing/delete/:outgoingID", DeleteOutgoing(dbh))

	req, _ := http.NewRequest(
		"POST", fmt.Sprintf("/outgoing/delete/%d", *outgoings[0].ID), nil,
	)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedResponse := `{"message":"Outgoing deleted!"}`
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponse)
	}
}
