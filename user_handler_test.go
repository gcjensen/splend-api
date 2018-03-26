package main

import (
	"database/sql"
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

	coupleID := InsertTestCouple(dbh)
	// Inserted so the partner of Hank can be tested

	InsertTestUser(&user.User{
		FirstName: "Marie",
		LastName:  "Schrader",
		Email:     "marie@schrader.com",
	}, coupleID, dbh)

	newUser := &user.User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	id := InsertTestUser(newUser, coupleID, dbh)

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

	DeleteAllUsers(dbh)
}

/***************** Methods for creating data to test against ******************/

func InsertTestUser(user *user.User, coupleID int, dbh *sql.DB) int {
	statement := fmt.Sprintf(`
        INSERT INTO users
        (first_name, last_name, email, couple_id)
        VALUES ("%s", "%s", "%s", %d)`,
		user.FirstName, user.LastName, user.Email, coupleID)

	dbh.Exec(statement)

	var id int
	dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

func InsertTestCouple(dbh *sql.DB) int {
	statement := fmt.Sprintf(
		`INSERT INTO couples (joining_date) VALUES ("2018-01-01")`,
	)

	dbh.Exec(statement)

	var id int
	dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

func DeleteAllUsers(dbh *sql.DB) {
	dbh.Exec("DELETE FROM users")
	dbh.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
	dbh.Exec("DELETE FROM couples")
	dbh.Exec("ALTER TABLE couples AUTO_INCREMENT = 1")
}
