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

func TestAddOutgoingFromMonzo(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(randomUser(), randomSha256(), dbh)

	router := httprouter.New()
	router.POST("/user/:id/monzo-webhook", AddOutgoingFromMonzo(dbh))

	bodyString := fmt.Sprintf(`{` +
		`"type":"transaction.created",` +
		`"data":{` +
		`"amount":-5432,` +
		`"description":"Aldi shop"` +
		`}}`)

	body := []byte(bodyString)

	id := strconv.Itoa(*user.ID)
	req, _ := http.NewRequest("POST", "/user/"+id+"/monzo-webhook", bytes.NewBuffer(body))

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

	assert.Equal(t, outgoings[0].Description, "Aldi shop")
	assert.Equal(t, outgoings[0].Amount, 5432)
}
