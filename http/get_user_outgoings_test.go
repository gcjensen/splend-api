package http

import (
	"fmt"
	"github.com/gcjensen/splend-api"
	"github.com/gcjensen/splend-api/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetUserOutgoingsEndPoint(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(randomUser(), randomSha256(), dbh)
	user.AddOutgoing(randomOutgoing())
	outgoings, _ := user.GetOutgoings()
	outgoing := outgoings[0]

	router := httprouter.New()
	router.GET("/user/:id/outgoings", GetUserOutgoings(dbh))

	url := fmt.Sprintf("/user/%d/outgoings", *user.ID)
	req, _ := http.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	splitTime := strings.Split(outgoing.Timestamp.String(), " ")
	timestamp := splitTime[0] + "T" + splitTime[1] + "Z"

	expected := fmt.Sprintf(`[{`+
		`"id":%d,`+
		`"description":"`+outgoing.Description+`",`+
		`"amount":"%d",`+
		`"owed":"%d",`+
		`"spender":"%d",`+
		`"category":"`+outgoing.Category+`",`+
		`"settled":null,`+
		`"timestamp":"`+timestamp+`"`+
		`}]`, *outgoing.ID, outgoing.Amount, outgoing.Owed, *user.ID)

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
