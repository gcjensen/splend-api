package http

import (
	"bytes"
	"github.com/gcjensen/splend-api"
	"github.com/gcjensen/splend-api/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestAddMonzoTransaction(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(randomUser(), randomSha256(), dbh)

	account := "acc_XXXXXXXXXXXXXXXXXXXXXX"
	user.LinkAccounts(&splend.LinkedAccounts{&account})

	router := httprouter.New()
	router.POST("/user/:id/monzo-webhook", AddFromMonzo(dbh))

	json, err := ioutil.ReadFile("../test/monzo-transaction.json")
	if err != nil {
		t.Errorf("Error loading test JSON file: %s", err)
	}

	body := []byte(json)

	id := strconv.Itoa(*user.ID)
	req, _ := http.NewRequest(
		"POST",
		"/user/"+id+"/monzo-webhook",
		bytes.NewBuffer(body),
	)

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

	outgoings, _ := user.GetOutgoings()

	assert.Equal(t, outgoings[0].Description, "Waitrose & Partners")
	assert.Equal(t, outgoings[0].Amount, 1254)
}
