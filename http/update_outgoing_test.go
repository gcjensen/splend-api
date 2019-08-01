package http

import (
	"bytes"
	"fmt"
	"github.com/gcjensen/splend-api"
	"github.com/gcjensen/splend-api/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(randomUser(), randomSha256(), dbh)
	user.AddOutgoing(randomOutgoing())
	outgoings, _ := user.GetOutgoings()
	outgoing := outgoings[0]

	router := httprouter.New()
	router.POST("/outgoing/update/:outgoingID", UpdateOutgoing(dbh))

	bodyString := fmt.Sprintf(`{`+
		`"description":"Groceries",`+
		`"amount":"60",`+
		`"owed":"30",`+
		`"spender":"%d",`+
		`"category":"General"`+
		`}`, *user.ID)

	body := []byte(bodyString)

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("/outgoing/update/%d", *outgoing.ID),
		bytes.NewBuffer(body),
	)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedResponse := `{"message":"Outgoing updated!"}`
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponse)
	}
}
