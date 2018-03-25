package main

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	DB := ResetDB()

	email := "jesse@pinkman.com"
	_, err := New(email, DB)

	assert.Equal(t, err.Error(), "User creation not yet implemented")
}

func TestNewFromDB(t *testing.T) {
	DB := ResetDB()

	coupleID := InsertTestCouple(DB)
	// Inserted so the partner of Hank can be tested
	InsertTestUser(&User{
		FirstName: "Marie",
		LastName:  "Schrader",
		Email:     "marie@schrader.com",
	}, coupleID, DB)

	newUser := &User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	id := InsertTestUser(newUser, coupleID, DB)

	user, err := NewFromDB(id, DB)

	assert.Nil(t, err)
	newUser.ID = id
	newUser.Partner = "Marie"
	assert.Equal(t, user, newUser)

	user, err = NewFromDB(10000, DB)
	assert.NotNil(t, err)
}

/***************** Methods for creating data to test against ******************/

func InsertTestUser(user *User, coupleID int, DB *sql.DB) int {
	statement := fmt.Sprintf(`
		INSERT INTO users
		(first_name, last_name, email, couple_id)
		VALUES ("%s", "%s", "%s", %d)`,
		user.FirstName, user.LastName, user.Email, coupleID)

	DB.Exec(statement)

	var id int
	DB.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

func InsertTestCouple(DB *sql.DB) int {
	statement := fmt.Sprintf(
		`INSERT INTO couples (joining_date) VALUES ("2018-01-01")`,
	)

	DB.Exec(statement)

	var id int
	DB.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}
