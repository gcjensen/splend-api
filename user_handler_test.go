package main

import (
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"github.com/gcjensen/settle-api/user"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGetDetailsEndPoint(t *testing.T) {
	dbh := config.TestDBH()
	server := Server{}
	server.Initialise(dbh)

	coupleID := config.InsertTestCouple(dbh)

	// Inserted so the partner of Hank can be tested
	partner := &user.User{
		FirstName: "Marie",
		LastName:  "Schrader",
		Email:     "marie@schrader.com",
	}
	config.InsertTestUser(
		partner.FirstName,
		partner.LastName,
		partner.Email,
		coupleID,
		dbh,
	)

	newUser := &user.User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	id := config.InsertTestUser(
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		coupleID,
		dbh,
	)

	req, _ := http.NewRequest("GET", "/user/{id}/details", nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(id)})

	rr := httptest.NewRecorder()
	userHandler := &UserHandler{dbh}
	handler := http.HandlerFunc(userHandler.GetDetails)

	handler.ServeHTTP(rr, req)

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
		`"partner":{"id":1,"name":"Marie"}`+
		`}`, id)

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	config.DeleteAllData(dbh)
}
