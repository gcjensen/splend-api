package user

import (
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
	config.DeleteAllData(dbh)
}

func TestNewFromDB(t *testing.T) {
	dbh := config.TestDBH()

	coupleID := config.InsertTestCouple(dbh)

	// Inserted so the partner of Hank can be tested
	partner := &User{
		FirstName: "Marie",
		LastName:  "Schrader",
		Email:     "marie@schrader.com",
	}
	partnerID := config.InsertTestUser(
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
	id := config.InsertTestUser(
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		coupleID,
		dbh,
	)

	user, err := NewFromDB(id, dbh)

	assert.Nil(t, err)
	newUser.ID = id
	newUser.Partner.Name = "Marie"
	newUser.Partner.ID = partnerID
	newUser.DBH = dbh
	assert.Equal(t, user, newUser)

	user, err = NewFromDB(10000, dbh)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Unknown user")

	config.DeleteAllData(dbh)
}

func TestGetOutgoings(t *testing.T) {

	dbh := config.TestDBH()

	coupleID := config.InsertTestCouple(dbh)
	newUser := &User{
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

	str := "2018-01-07T15:32:12.000Z"
	timestamp, err := time.Parse(time.RFC3339, str)
	newOutgoing := outgoing.Outgoing{
		0, "Minerals", 200.00, 10.00, id, "General", nil, timestamp,
	}
	outgoingID := config.InsertTestOutgoing(
		newOutgoing.Description,
		newOutgoing.Amount,
		newOutgoing.Owed,
		newOutgoing.Spender,
		newOutgoing.Timestamp,
		dbh,
	)
	newOutgoing.ID = outgoingID

	user, err := NewFromDB(id, dbh)
	outgoings, err := user.GetOutgoings()

	assert.Equal(t, []outgoing.Outgoing{newOutgoing}, outgoings)
	assert.Nil(t, err)

	config.DeleteAllData(dbh)
}
