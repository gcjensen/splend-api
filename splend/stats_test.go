package splend_test

import (
	"testing"
	"time"

	"github.com/gcjensen/splend-api/config"
	"github.com/gcjensen/splend-api/splend"
	"github.com/gcjensen/splend-api/test"
	"github.com/stretchr/testify/assert"
)

func TestUser_GetMonthBreakdown(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)
	randomPartner := test.RandomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, _ := splend.NewUser(randomPartner, test.RandomSha256(), dbh)
	user.Partner = partner

	groceries := &splend.Outgoing{
		Description: "Weekly shop",
		Amount:      5000,
		Owed:        2500,
		Category:    "Groceries",
	}

	groceriesAgain := &splend.Outgoing{
		Description: "Weekly shop",
		Amount:      4000,
		Owed:        2000,
		Category:    "Groceries",
	}

	beers := &splend.Outgoing{
		Description: "Beers",
		Amount:      2000,
		Owed:        0,
		Category:    "Drinks",
	}

	_ = user.AddOutgoing(groceries)
	_ = partner.AddOutgoing(groceriesAgain)
	_ = user.AddOutgoing(beers)

	breakdown, err := user.GetMonthBreakdown(time.Now().Format("2006-01"))
	assert.Nil(t, err)
	assert.Equal(
		t, []splend.CategoryTotals{
			{"Groceries", 4500, 9000},
			{"Drinks", 2000, 0},
		}, breakdown,
	)
}
