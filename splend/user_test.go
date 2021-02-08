package splend_test

import (
	"errors"
	"testing"

	"github.com/gcjensen/amex"
	"github.com/gcjensen/splend-api/config"
	"github.com/gcjensen/splend-api/splend"
	"github.com/gcjensen/splend-api/test"
	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)

	randomPartner := test.RandomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, _ := splend.NewUser(randomPartner, test.RandomSha256(), dbh)

	user.Partner = partner

	userFromDB, err := splend.NewUserFromDB(*user.ID, dbh)

	assert.Nil(t, err)
	assert.Equal(t, user.ID, userFromDB.ID)
	assert.Equal(t, user.Email, userFromDB.Email)

	_, err = splend.NewUserFromDB(10000, dbh)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, splend.ErrUserUnknown))
}

func TestUser_AddGetOutgoings(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)
	randomPartner := test.RandomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, _ := splend.NewUser(randomPartner, test.RandomSha256(), dbh)
	user.Partner = partner

	randomOutgoingOne := test.RandomOutgoing()
	randomOutgoingTwo := test.RandomOutgoing()

	// Won't be included as it's the partner's outgoing and owed is 0
	randomOutgoingThree := &splend.Outgoing{
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

	where := map[string]interface{}{}
	outgoings, err := user.GetOutgoings(where)

	// Time of insertion is used (so hard to mock), so we just manually set
	// it here
	randomOutgoingOne.Timestamp = outgoings[0].Timestamp
	randomOutgoingTwo.Timestamp = outgoings[1].Timestamp

	assert.ElementsMatch(t,
		[]string{randomOutgoingOne.Description, randomOutgoingTwo.Description},
		[]string{outgoings[0].Description, outgoings[1].Description},
	)
	assert.Nil(t, err)

	where["description"] = randomOutgoingOne.Description

	outgoings, _ = user.GetOutgoings(where)
	assert.Len(t, outgoings, 1)
}

func TestUser_GetSummary(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)

	outgoings := []*splend.Outgoing{
		test.RandomOutgoing(),
		test.RandomOutgoing(),
		test.RandomOutgoing(),
	}

	owed := 0
	for _, o := range outgoings {
		owed += o.Owed
		err := user.AddOutgoing(o)
		assert.Nil(t, err)
	}

	s, err := user.GetSummary()

	assert.Equal(t, owed, s.Balance)
	assert.Nil(t, err)
}

func TestUser_AddAmexTransaction(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)

	amexTX := amex.Transaction{
		Amount:      1400,
		Date:        "01-01-20",
		Description: "Beers",
		ID:          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	err := user.AddAmexTransaction(amexTX)
	assert.Nil(t, err)

	outgoings, _ := user.GetOutgoings(nil)
	assert.Equal(t, amexTX.Amount, outgoings[0].Amount)
	assert.Equal(t, amexTX.Description, outgoings[0].Description)

	err = user.AddAmexTransaction(amexTX)
	assert.True(t, errors.Is(err, splend.ErrAlreadyExists))
}
