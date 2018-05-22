package endpoints

import (
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"github.com/gcjensen/settle-api/outgoing"
	"github.com/gcjensen/settle-api/user"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetUserOutgoingsEndPoint(t *testing.T) {
	dbh := config.TestDBH()

	coupleID := config.InsertTestCouple(dbh)

	newUser := &user.User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	userID := config.InsertTestUser(
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		coupleID,
		dbh,
	)

	timestamp, _ := time.Parse(time.RFC3339, "2018-01-07T15:32:12.000Z")
	newOutgoing := outgoing.Outgoing{
		nil, "Minerals", 200.00, 10.00, userID, "General", nil, &timestamp,
	}
	outgoingID := config.InsertTestOutgoing(
		newOutgoing.Description,
		newOutgoing.Amount,
		newOutgoing.Owed,
		newOutgoing.Spender,
		*newOutgoing.Timestamp,
		dbh,
	)

	router := httprouter.New()
	router.GET("/user/:id/outgoings", GetUserOutgoings(dbh))

	req, _ := http.NewRequest("GET", fmt.Sprintf("/user/%d/outgoings", userID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := fmt.Sprintf(`[{`+
		`"id":%d,`+
		`"description":"Minerals",`+
		`"amount":"200",`+
		`"owed":"10",`+
		`"spender":"%d",`+
		`"category":"General",`+
		`"settled":null,`+
		`"timestamp":"2018-01-07T15:32:12Z"`+
		`}]`, outgoingID, userID)

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	config.DeleteAllData(dbh)
}
