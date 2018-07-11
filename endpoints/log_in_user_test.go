package endpoints

import (
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"github.com/gcjensen/settle-api/test"
	"github.com/gcjensen/settle-api/user"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogInUser(t *testing.T) {
	dbh := config.TestDBH()

	coupleID := test.InsertTestCouple(dbh)

	colour := "FFFFFF"

	// Inserted so the partner of Hank can be tested
	partner := &user.User{
		FirstName: "Marie",
		LastName:  "Schrader",
		Email:     "marie@schrader.com",
		Colour:    &colour,
	}
	test.InsertTestUser(
		partner.FirstName,
		partner.LastName,
		partner.Email,
		*partner.Colour,
		coupleID,
		dbh,
	)

	newUser := &user.User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
		Colour:    &colour,
	}
	id := test.InsertTestUser(
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		*newUser.Colour,
		coupleID,
		dbh,
	)

	router := httprouter.New()
	router.POST("/user/:id", LogInUser(dbh))

	req, _ := http.NewRequest("POST", "/user/2", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := fmt.Sprintf(`{`+
		`"id":%d,`+
		`"firstName":"Hank",`+
		`"lastName":"Schrader",`+
		`"email":"hank@schrader.com",`+
		`"colour":"FFFFFF",`+
		`"partner":{"id":1,"name":"Marie","colour":"FFFFFF"}`+
		`}`, id)

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	test.DeleteAllData(dbh)
}
