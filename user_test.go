package splend

import (
	"github.com/gcjensen/splend/config"
	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAndNewFromDB(t *testing.T) {
	dbh := config.TestDBH()

	user, err := NewUser(randomUser(), dbh)

	randomPartner := randomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, err := NewUser(randomPartner, dbh)
	partner.dbh = nil

	partner.Partner = nil
	user.Partner = partner

	userFromDB, err := NewUserFromDB(*user.ID, dbh)

	assert.Nil(t, err)
	assert.Equal(t, user, userFromDB)

	user, err = NewUserFromDB(10000, dbh)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Unknown user")
}

func TestAddAndGetOutgoings(t *testing.T) {
	dbh := config.TestDBH()

	user, err := NewUser(randomUser(), dbh)
	randomPartner := randomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, err := NewUser(randomPartner, dbh)
	user.Partner = partner

	randomOutgoingOne := randomOutgoing()
	randomOutgoingTwo := randomOutgoing()

	err = user.AddOutgoing(randomOutgoingOne)
	assert.Nil(t, err)
	err = partner.AddOutgoing(randomOutgoingTwo)
	assert.Nil(t, err)

	outgoings, err := user.GetOutgoings()

	// Time of insertion is used (so hard to mock), so we just manually set
	// it here
	randomOutgoingOne.Timestamp = outgoings[0].Timestamp
	randomOutgoingTwo.Timestamp = outgoings[1].Timestamp

	assert.Equal(
		t, []Outgoing{*randomOutgoingOne, *randomOutgoingTwo}, outgoings,
	)
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