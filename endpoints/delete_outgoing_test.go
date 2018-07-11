package endpoints

import (
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"github.com/gcjensen/settle-api/outgoing"
	"github.com/gcjensen/settle-api/test"
	"github.com/gcjensen/settle-api/user"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeleteOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	coupleID := test.InsertTestCouple(dbh)

	newUser := &user.User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	userID := test.InsertTestUser(
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		"",
		coupleID,
		dbh,
	)

	timestamp, _ := time.Parse(time.RFC3339, "2018-01-07T15:32:12.000Z")
	newOutgoing := outgoing.Outgoing{
		Description: "Minerals", Amount: 200.00, Owed: 10.00, Spender: userID,
		Category: "General", Timestamp: &timestamp,
	}
	outgoingID := test.InsertTestOutgoing(
		newOutgoing.Description,
		newOutgoing.Amount,
		newOutgoing.Owed,
		newOutgoing.Spender,
		*newOutgoing.Timestamp,
		dbh,
	)

	router := httprouter.New()
	router.POST("/outgoing/delete/:outgoingID", DeleteOutgoing(dbh))

	req, _ := http.NewRequest(
		"POST", fmt.Sprintf("/outgoing/delete/%d", outgoingID), nil,
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

	test.DeleteAllData(dbh)
}
