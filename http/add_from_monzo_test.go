package http

import (
	"bytes"
	"fmt"
	"github.com/gcjensen/splend-api"
	"github.com/gcjensen/splend-api/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
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

	bodyString := fmt.Sprintf(`{` +
		`"type":"transaction.created",` +
		`"data": {` +
		`"account_id": "` + account + `",` +
		`"amount": -5432,` +
		`"description": "Aldi shop",` +
		`"merchant": {` +
		`"name": "Aldi"` +
		`}}}`)

	body := []byte(bodyString)

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

	assert.Equal(t, outgoings[0].Description, "Aldi")
	assert.Equal(t, outgoings[0].Amount, 5432)
}