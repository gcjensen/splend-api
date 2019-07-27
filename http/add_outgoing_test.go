package http

import (
	"bytes"
	"fmt"
	"github.com/gcjensen/splend"
	"github.com/gcjensen/splend/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestAddOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(randomUser(), randomSha256(), dbh)

	router := httprouter.New()
	router.POST("/user/:id/add", AddOutgoing(dbh))

	bodyString := fmt.Sprintf(`{`+
		`"description":"Minerals",`+
		`"amount":"200",`+
		`"owed":"10",`+
		`"spender":"%d",`+
		`"category":"General"`+
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
