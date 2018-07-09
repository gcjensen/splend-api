package outgoing

import (
	"database/sql"
	"github.com/gcjensen/settle-api/config"
	"github.com/gcjensen/settle-api/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	dbh := config.TestDBH()
	testOutgoing := insertTestOutgoing(dbh)

	outgoing, err := New(*testOutgoing.ID, dbh)

	assert.Nil(t, err)
	assert.Equal(t, &testOutgoing, outgoing)

	outgoing, err = New(10000, dbh)
	assert.NotNil(t, err)
	assert.Equal(t, "Proper outgoing creation not yet implemented", err.Error())

	test.DeleteAllData(dbh)
}

func TestDelete(t *testing.T) {
	dbh := config.TestDBH()
	testOutgoing := insertTestOutgoing(dbh)

	err := testOutgoing.Delete()

	count := test.GetOutgoingCount(dbh)

	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	test.DeleteAllData(dbh)
}

func TestToggleSettled(t *testing.T) {
	dbh := config.TestDBH()
	testOutgoing := insertTestOutgoing(dbh)

	outgoing, _ := New(*testOutgoing.ID, dbh)

	assert.Nil(t, outgoing.Settled)

	err := outgoing.ToggleSettled(true)

	assert.Nil(t, err)
	assert.NotNil(t, outgoing.Settled)

	test.DeleteAllData(dbh)
}

/************************** Private Implementation ****************************/

/*
 * Uses the functions in the test package to insert an outgoing (and the
 * required user)
 */
func insertTestOutgoing(dbh *sql.DB) Outgoing {
	coupleID := test.InsertTestCouple(dbh)
	userID := test.InsertTestUser(
		"Wade", "Wilson", "wade@wilson.com", coupleID, dbh,
	)

	str := "2018-01-07T15:32:12.000Z"
	timestamp, _ := time.Parse(time.RFC3339, str)
	testOutgoing := Outgoing{
		nil, "New suit", 200.00, 10.00, userID, "General", nil, &timestamp, dbh,
	}
	outgoingID := test.InsertTestOutgoing(
		testOutgoing.Description,
		testOutgoing.Amount,
		testOutgoing.Owed,
		testOutgoing.Spender,
		*testOutgoing.Timestamp,
		dbh,
	)
	testOutgoing.ID = &outgoingID

	return testOutgoing
}
