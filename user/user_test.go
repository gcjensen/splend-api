package user

import (
	"database/sql"
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	dbh := config.TestDBH()

	email := "jesse@pinkman.com"
	_, err := New(email, dbh)

	assert.Equal(t, err.Error(), "User creation not yet implemented")
	DeleteAllUsers(dbh)
}

func TestNewFromDB(t *testing.T) {
	dbh := config.TestDBH()

	coupleID := InsertTestCouple(dbh)

	// Inserted so the partner of Hank can be tested
	InsertTestUser(&User{
		FirstName: "Marie",
		LastName:  "Schrader",
		Email:     "marie@schrader.com",
	}, coupleID, dbh)

	newUser := &User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	id := InsertTestUser(newUser, coupleID, dbh)

	user, err := NewFromDB(id, dbh)

	assert.Nil(t, err)
	newUser.ID = id
	newUser.Partner = "Marie"
	assert.Equal(t, user, newUser)

	user, err = NewFromDB(10000, dbh)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Unknown user")

	DeleteAllUsers(dbh)
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

func DeleteAllUsers(DB *sql.DB) {
	DB.Exec("DELETE FROM users")
}
