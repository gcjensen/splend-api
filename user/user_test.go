package user

import (
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"github.com/gcjensen/settle-api/outgoing"
	"github.com/gcjensen/settle-api/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	dbh := config.TestDBH()

	email := "jesse@pinkman.com"
	_, err := New(email, dbh)

	assert.Equal(t, err.Error(), "User creation not yet implemented")
	test.DeleteAllData(dbh)
}

func TestNewFromDB(t *testing.T) {
	dbh := config.TestDBH()

	coupleID := test.InsertTestCouple(dbh)

	// Inserted so the partner of Hank can be tested
	partner := &User{
		FirstName: "Marie",
		LastName:  "Schrader",
		Email:     "marie@schrader.com",
	}
	partnerID := test.InsertTestUser(
		partner.FirstName,
		partner.LastName,
		partner.Email,
		coupleID,
		dbh,
	)

	newUser := &User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	id := test.InsertTestUser(
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		coupleID,
		dbh,
	)

	user, err := NewFromDB(id, dbh)

	assert.Nil(t, err)
	newUser.ID = &id
	newUser.Partner.Name = "Marie"
	newUser.dbh = dbh
	newUser.Partner.ID = partnerID
	assert.Equal(t, user, newUser)

	user, err = NewFromDB(10000, dbh)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Unknown user")

	test.DeleteAllData(dbh)
}

func TestGetOutgoings(t *testing.T) {

	dbh := config.TestDBH()

	coupleID := test.InsertTestCouple(dbh)
	newUser := &User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	id := test.InsertTestUser(
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		coupleID,
		dbh,
	)

	str := "2018-01-07T15:32:12.000Z"
	timestamp, err := time.Parse(time.RFC3339, str)
	newOutgoing := outgoing.Outgoing{
		Description: "Minerals", Amount: 200.00, Owed: 10.00, Spender: id,
		Category: "General", Timestamp: &timestamp,
	}
	outgoingID := test.InsertTestOutgoing(
		newOutgoing.Description,
		newOutgoing.Amount,
		newOutgoing.Owed,
		newOutgoing.Spender,
		*newOutgoing.Timestamp,
		dbh,
	)
	newOutgoing.ID = &outgoingID

	user, err := NewFromDB(id, dbh)
	outgoings, err := user.GetOutgoings()

	assert.Equal(t, []outgoing.Outgoing{newOutgoing}, outgoings)
	assert.Nil(t, err)

	test.DeleteAllData(dbh)
}

func TestAddOutgoings(t *testing.T) {

	dbh := config.TestDBH()

	coupleID := test.InsertTestCouple(dbh)
	newUser := &User{
		FirstName: "Hank",
		LastName:  "Schrader",
		Email:     "hank@schrader.com",
	}
	id := test.InsertTestUser(
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		coupleID,
		dbh,
	)

	newOutgoing := outgoing.Outgoing{
		Description: "Fried chicken", Amount: 7.00, Owed: 3.5, Spender: id,
		Category: "General",
	}

	user, err := NewFromDB(id, dbh)
	err = user.AddOutgoing(newOutgoing)

	statement := fmt.Sprintf(
		`SELECT description, spender_id FROM outgoings LIMIT 1`,
	)

	var description string
	var spenderID int
	err = dbh.QueryRow(statement).Scan(&description, &spenderID)

	assert.Equal(t, description, newOutgoing.Description)
	assert.Nil(t, err)

	test.DeleteAllData(dbh)
}
