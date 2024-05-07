package api_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gcjensen/splend-api/api"
	"github.com/gcjensen/splend-api/config"
	"github.com/gcjensen/splend-api/splend"
	"github.com/gcjensen/splend-api/test"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestAddAmexTransaction(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)

	router := httprouter.New()
	router.POST("/user/:id/amex", api.AddFromAmex(dbh))

	bodyString := `{
		"amount":"1700",
		"date":"2020-01-01T19:30:31Z",
		"description":"Beers",
		"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	}`

	body := []byte(bodyString)

	id := strconv.Itoa(*user.ID)
	req, _ := http.NewRequest("POST", "/user/"+id+"/amex", bytes.NewBuffer(body))

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedResponse := `{"message":"Amex transaction added"}`
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponse)
	}

	outgoings, _ := user.GetOutgoings(nil)

	assert.Equal(t, "Beers", outgoings[0].Description)
	assert.Equal(t, 1700, outgoings[0].Amount)
	assert.Equal(t, time.Date(2020, 01, 01, 19, 30, 31, 0, time.UTC), *outgoings[0].Timestamp)
}

func TestAddMonzoTransaction(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)

	account := "acc_XXXXXXXXXXXXXXXXXXXXXX"
	err := user.LinkMonzo(&splend.MonzoAccount{ID: &account})
	if err != nil {
		t.Error(err)
	}

	jointAccount := "acc_YYYYYYYYYYYYYYYYYYYYYY"
	err = user.LinkMonzo(&splend.MonzoAccount{ID: &jointAccount})
	if err != nil {
		t.Error(err)
	}

	router := httprouter.New()
	router.POST("/user/:id/monzo-webhook", api.AddFromMonzo(dbh))

	json, err := ioutil.ReadFile("../test/monzo-transaction.json")
	if err != nil {
		t.Errorf("Error loading test JSON file: %s", err)
	}

	id := strconv.Itoa(*user.ID)
	req, _ := http.NewRequest("POST", "/user/"+id+"/monzo-webhook", bytes.NewBuffer(json))

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedResponse := `{"message":"Request successful"}`
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponse)
	}

	outgoings, _ := user.GetOutgoings(nil)

	assert.Equal(t, "Waitrose & Partners", outgoings[0].Description)
	assert.Equal(t, 1254, outgoings[0].Amount)
}
