package user

import (
	"github.com/gcjensen/settle-api/config"
	"github.com/gcjensen/settle-api/outgoing"
	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
)

func TestNewAndNewFromDB(t *testing.T) {
	dbh := config.TestDBH()

	user, err := New(randomUser(), dbh)

	randomPartner := randomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, err := New(randomPartner, dbh)
	partner.dbh = nil

	partner.Partner = nil
	user.Partner = partner

	userFromDB, err := NewFromDB(*user.ID, dbh)

	assert.Nil(t, err)
	assert.Equal(t, user, userFromDB)

	user, err = NewFromDB(10000, dbh)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Unknown user")
}

func TestAddAndGetOutgoings(t *testing.T) {
	dbh := config.TestDBH()

	user, err := New(randomUser(), dbh)
	randomOutgoing := randomOutgoing()

	err = user.AddOutgoing(randomOutgoing)
	assert.Nil(t, err)

	outgoings, err := user.GetOutgoings()

	// Time of insertion is used (so hard to mock), so we just manually set
	// it here
	randomOutgoing.Timestamp = outgoings[0].Timestamp

	assert.Equal(t, []outgoing.Outgoing{*randomOutgoing}, outgoings)
	assert.Nil(t, err)
}

/***************************** Test data insertion ****************************/

func randomUser() *User {
	colour := "FFFFFF"
	return &User{
		FirstName: fake.FirstName(),
		LastName:  fake.LastName(),
		Email:     fake.EmailAddress(),
		Colour:    &colour,
	}
}

func randomOutgoing() *outgoing.Outgoing {
	amount := math.Ceil(rand.Float64()*100) / 100
	return &outgoing.Outgoing{
		Description: fake.ProductName(),
		Amount:      amount,
		Owed:        amount / 2,
		Category:    fake.Product(),
	}
}
