package user

import (
	"database/sql"
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"github.com/gcjensen/settle-api/outgoing"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	dbh := config.TestDBH()

	email := "jesse@pinkman.com"
	_, err := New(email, dbh)

	assert.Equal(t, err.Error(), "User creation not yet implemented")
	DeleteAllData(dbh)
}

func TestNewFromDB(t *testing.T) {
	dbh := config.TestDBH()

	coupleID := InsertTestCouple(dbh)

	// Inserted so the partner of Hank can be tested
	partnerID := InsertTestUser(&User{
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
	newUser.Partner.Name = "Marie"
	newUser.Partner.ID = partnerID
	assert.Equal(t, user, newUser)

	user, err = NewFromDB(10000, dbh)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Unknown user")

	DeleteAllData(dbh)
}

func TestGetOutgoings(t *testing.T) {

	dbh := config.TestDBH()

	coupleID := InsertTestCouple(dbh)
	newUser := &User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	id := InsertTestUser(newUser, coupleID, dbh)

	str := "2018-01-07T15:32:12.000Z"
	timestamp, err := time.Parse(time.RFC3339, str)
	testOutgoing := outgoing.Outgoing{
		0, "Minerals", 200.00, id, "General", nil, timestamp,
	}
	outgoingID := InsertTestOutgoing(&testOutgoing, dbh)
	testOutgoing.ID = outgoingID

	user, err := NewFromDB(id, dbh)
	outgoings, err := user.GetOutgoings(dbh)

	assert.Equal(t, []outgoing.Outgoing{testOutgoing}, outgoings)
	assert.Nil(t, err)

	DeleteAllData(dbh)
}

/***************** Methods for creating data to test against ******************/

func InsertTestUser(user *User, coupleID int, dbh *sql.DB) int {
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

func InsertTestOutgoing(o *outgoing.Outgoing, dbh *sql.DB) int {
	statement := `INSERT INTO categories (id, name) VALUES (1, "General")`
	dbh.Exec(statement)

	statement = fmt.Sprintf(`
		INSERT INTO outgoings
		(description, amount, spender_id, category_id, settled, timestamp)
		VALUES ("%s", %f, %d, %d, NULL, "%s")`,
		o.Description, o.Amount, o.Spender, 1, o.Timestamp)

	dbh.Exec(statement)

	var id int
	dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

func DeleteAllData(dbh *sql.DB) {
	dbh.Exec("DELETE FROM outgoings")
	dbh.Exec("ALTER TABLE outgoings AUTO_INCREMENT = 1")
	dbh.Exec("DELETE FROM users")
	dbh.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
	dbh.Exec("DELETE FROM couples")
	dbh.Exec("ALTER TABLE couples AUTO_INCREMENT = 1")
}
