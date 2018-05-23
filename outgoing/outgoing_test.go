package outgoing

import (
	"github.com/gcjensen/settle-api/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	dbh := config.TestDBH()

	coupleID := config.InsertTestCouple(dbh)
	userID := config.InsertTestUser(
		"Wade", "Wilson", "wade@wilson.com", coupleID, dbh,
	)

	str := "2018-01-07T15:32:12.000Z"
	timestamp, err := time.Parse(time.RFC3339, str)
	testOutgoing := Outgoing{
		nil, "New suit", 200.00, 10.00, userID, "General", nil, &timestamp, dbh,
	}
	outgoingID := config.InsertTestOutgoing(
		testOutgoing.Description,
		testOutgoing.Amount,
		testOutgoing.Owed,
		testOutgoing.Spender,
		*testOutgoing.Timestamp,
		dbh,
	)
	testOutgoing.ID = &outgoingID

	outgoing, err := New(outgoingID, dbh)

	assert.Nil(t, err)
	assert.Equal(t, &testOutgoing, outgoing)

	outgoing, err = New(10000, dbh)
	assert.NotNil(t, err)
	assert.Equal(t, "Proper outgoing creation not yet implemented", err.Error())

	config.DeleteAllData(dbh)
}

func TestToggleSettled(t *testing.T) {
	dbh := config.TestDBH()

	coupleID := config.InsertTestCouple(dbh)
	userID := config.InsertTestUser(
		"Wade", "Wilson", "wade@wilson.com", coupleID, dbh,
	)

	str := "2018-01-07T15:32:12.000Z"
	timestamp, err := time.Parse(time.RFC3339, str)
	testOutgoing := Outgoing{
		nil, "New suit", 200.00, 10.00, userID, "General", nil, &timestamp, dbh,
	}
	outgoingID := config.InsertTestOutgoing(
		testOutgoing.Description,
		testOutgoing.Amount,
		testOutgoing.Owed,
		testOutgoing.Spender,
		*testOutgoing.Timestamp,
		dbh,
	)
	testOutgoing.ID = &outgoingID

	outgoing, _ := New(outgoingID, dbh)

	assert.Nil(t, outgoing.Settled)

	err = outgoing.ToggleSettled(true)

	assert.Nil(t, err)
	assert.NotNil(t, outgoing.Settled)

	config.DeleteAllData(dbh)
}
