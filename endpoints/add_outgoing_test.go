package endpoints

import (
	"bytes"
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"github.com/gcjensen/settle-api/user"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestAddOutgoing(t *testing.T) {
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

	router := httprouter.New()
	router.POST("/user/:id/add", AddOutgoing(dbh))

	bodyString := fmt.Sprintf(`{`+
		`"description":"Minerals",`+
		`"amount":"200",`+
		`"owed":"10",`+
		`"spender":"%d",`+
		`"category":"General"`+
		`}`, userID)

	body := []byte(bodyString)

	id := strconv.Itoa(userID)
	req, _ := http.NewRequest("POST", "/user/"+id+"/add", bytes.NewBuffer(body))

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	expectedResponse := `{"message":"Outgoing added!"}`
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponse)
	}

	config.DeleteAllData(dbh)
}