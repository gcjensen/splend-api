package api_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gcjensen/splend-api/api"
	"github.com/gcjensen/splend-api/config"
	"github.com/gcjensen/splend-api/splend"
	"github.com/gcjensen/splend-api/test"
	"github.com/julienschmidt/httprouter"
)

func TestAddOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)

	router := httprouter.New()
	router.POST("/user/:id/add", api.AddOutgoing(dbh))

	bodyString := fmt.Sprintf(`{`+
		`"description":"Minerals",`+
		`"amount":"200",`+
		`"owed":"10",`+
		`"spender":"%d",`+
		`"category":"General",`+
		`"tags":["discretionary"]`+
		`}`, *user.ID)

	body := []byte(bodyString)

	id := strconv.Itoa(*user.ID)
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
}

func TestDeleteOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)
	_ = user.AddOutgoing(test.RandomOutgoing())
	outgoings, _ := user.GetOutgoings(nil)

	router := httprouter.New()
	router.POST("/outgoing/delete/:outgoingID", api.DeleteOutgoing(dbh))

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

func TestSettleOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)
	_ = user.AddOutgoing(test.RandomOutgoing())
	outgoings, _ := user.GetOutgoings(nil)
	outgoing := outgoings[0]

	router := httprouter.New()
	router.POST(
		"/outgoing/settle/:outgoingID/:shouldSettle",
		api.SettleOutgoing(dbh),
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

func TestUpdateOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)
	_ = user.AddOutgoing(test.RandomOutgoing())
	outgoings, _ := user.GetOutgoings(nil)
	outgoing := outgoings[0]

	router := httprouter.New()
	router.POST("/outgoing/update/:outgoingID", api.UpdateOutgoing(dbh))

	bodyString := fmt.Sprintf(`{`+
		`"description":"Groceries",`+
		`"amount":"60",`+
		`"owed":"30",`+
		`"spender":"%d",`+
		`"category":"General",`+
		`"tags":["discretionary"]`+
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
