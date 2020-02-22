package splend

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"

	"github.com/gcjensen/amex"
	"github.com/gcjensen/splend-api/config"
	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
)

func TestNewAndNewFromDB(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := NewUser(randomUser(), randomSha256(), dbh)

	randomPartner := randomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, _ := NewUser(randomPartner, randomSha256(), dbh)
	partner.dbh = nil

	partner.Partner = nil
	user.Partner = partner

	userFromDB, err := NewUserFromDB(*user.ID, dbh)

	assert.Nil(t, err)
	assert.Equal(t, user, userFromDB)

	_, err = NewUserFromDB(10000, dbh)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "unknown user")
}

func TestAddAndGetOutgoings(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := NewUser(randomUser(), randomSha256(), dbh)
	randomPartner := randomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, _ := NewUser(randomPartner, randomSha256(), dbh)
	user.Partner = partner

	randomOutgoingOne := randomOutgoing()
	randomOutgoingTwo := randomOutgoing()

	// Won't be included as it's the partner's outgoing and owed is 0
	randomOutgoingThree := &Outgoing{
		Description: fake.ProductName(),
		Amount:      10,
		Owed:        0,
		Category:    fake.Product(),
	}

	err := user.AddOutgoing(randomOutgoingOne)
	assert.Nil(t, err)
	err = partner.AddOutgoing(randomOutgoingTwo)
	assert.Nil(t, err)
	err = partner.AddOutgoing(randomOutgoingThree)
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

func TestAddAmexTransaction(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := NewUser(randomUser(), randomSha256(), dbh)

	amexTX := amex.Transaction{
		Amount:      1400,
		Date:        "01-01-20",
		Description: "Beers",
		ID:          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	err := user.AddAmexTransaction(amexTX)
	assert.Nil(t, err)

	outgoings, _ := user.GetOutgoings()
	assert.Equal(t, amexTX.Amount, outgoings[0].Amount)
	assert.Equal(t, amexTX.Description, outgoings[0].Description)

	err = user.AddAmexTransaction(amexTX)
	assert.True(t, errors.Is(err, ErrAlreadyExists))
}

/***************************** Test data insertion ****************************/

func randomSha256() string {
	return fmt.Sprintf(
		"%x",
		sha256.Sum256([]byte(fake.Digits())),
	)
}

func randomUser() *User {
	colour := "FFFFFF"

	return &User{
		FirstName: fake.FirstName(),
		LastName:  fake.LastName(),
		Email:     fake.EmailAddress(),
		Colour:    &colour,
	}
}
