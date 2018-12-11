package endpoints

import (
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

func TestLogInUser(t *testing.T) {
	dbh := config.TestDBH()

	tempUser, _ := user.New(randomUser(), dbh)

	randomUser := randomUser()
	randomUser.CoupleID = tempUser.CoupleID
	testUser, _ := user.New(randomUser, dbh)

	router := httprouter.New()
	router.POST("/user/:id", LogInUser(dbh))

	id := strconv.Itoa(*testUser.ID)
	req, _ := http.NewRequest("POST", "/user/"+id, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := fmt.Sprintf(`{`+
		`"id":%d,`+
		`"firstName":"`+testUser.FirstName+`",`+
		`"lastName":"`+testUser.LastName+`",`+
		`"email":"`+testUser.Email+`",`+
		`"colour":"`+*testUser.Colour+`",`+
		`"partner":{`+
		`"id":%d,`+
		`"firstName":"`+testUser.Partner.FirstName+`",`+
		`"lastName":"`+testUser.Partner.LastName+`",`+
		`"email":"`+testUser.Partner.Email+`",`+
		`"colour":"`+*testUser.Partner.Colour+`",`+
		`"partner":null,`+
		`"iconLink":null},`+
		`"iconLink":null`+
		`}`, *testUser.ID, *testUser.Partner.ID)

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
