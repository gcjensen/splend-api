package splend

import (
	"github.com/gcjensen/splend-api/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetMonthBreakdown(t *testing.T) {
	dbh := config.TestDBH()

	user, err := NewUser(randomUser(), randomSha256(), dbh)
	randomPartner := randomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, err := NewUser(randomPartner, randomSha256(), dbh)
	user.Partner = partner

	groceries := &Outgoing{
		Description: "Weekly shop",
		Amount:      5000,
		Owed:        2500,
		Category:    "Groceries",
	}

	groceriesAgain := &Outgoing{
		Description: "Weekly shop",
		Amount:      4000,
		Owed:        2000,
		Category:    "Groceries",
	}

	beers := &Outgoing{
		Description: "Beers",
		Amount:      2000,
		Owed:        0,
		Category:    "Drinks",
	}

	_ = user.AddOutgoing(groceries)
	_ = partner.AddOutgoing(groceriesAgain)
	_ = user.AddOutgoing(beers)

	breakdown, err := user.GetMonthBreakdown(time.Now().Format("2006-01"), true)
	assert.Nil(t, err)
	assert.Equal(
		t, []CategoryTotal{
			CategoryTotal{"Groceries", 9000},
			CategoryTotal{"Drinks", 2000},
		}, breakdown,
	)

	breakdown, err = user.GetMonthBreakdown(time.Now().Format("2006-01"), false)
	assert.Nil(t, err)
	assert.Equal(
		t, []CategoryTotal{
			CategoryTotal{"Groceries", 4500},
			CategoryTotal{"Drinks", 2000},
		}, breakdown,
	)
}
