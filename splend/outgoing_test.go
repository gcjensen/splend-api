package splend_test

import (
	"errors"
	"testing"

	"github.com/gcjensen/splend-api/config"
	"github.com/gcjensen/splend-api/splend"
	"github.com/gcjensen/splend-api/test"
	"github.com/stretchr/testify/assert"
)

func TestNewOutgoing(t *testing.T) {
	dbh := config.TestDBH()

	randomOutgoing := test.RandomUserAndOutgoing(dbh)
	outgoing, _ := splend.NewOutgoing(randomOutgoing, dbh)

	outgoingFromDB, err := splend.NewOutgoingFromDB(*outgoing.ID, dbh)
	outgoing.Timestamp = outgoingFromDB.Timestamp

	assert.Nil(t, err)
	assert.Equal(t, outgoing, outgoingFromDB)

	_, err = splend.NewOutgoingFromDB(10000, dbh)
	assert.NotNil(t, err)
}

func TestOutgoing_Delete(t *testing.T) {
	dbh := config.TestDBH()
	randomOutgoing := test.RandomUserAndOutgoing(dbh)
	outgoing, _ := splend.NewOutgoing(randomOutgoing, dbh)

	err := outgoing.Delete()
	assert.Nil(t, err)

	_, err = splend.NewOutgoingFromDB(*outgoing.ID, dbh)

	assert.True(t, errors.Is(err, splend.ErrOutgoingUnknown))

	// Make sure we can also delete outgoings that exist in amex_transactions
	outgoing, _ = splend.NewOutgoing(test.RandomUserAndOutgoing(dbh), dbh)

	statement, _ := dbh.Prepare(`
		INSERT INTO amex_transactions
		(amex_id, outgoing_id)
		VALUES ("aaaaaaaaaaaaaaaaaaaaaaa", ?)
	`)
	defer statement.Close()

	_, _ = statement.Exec(outgoing.ID)

	err = outgoing.Delete()
	assert.Nil(t, err)

	_, err = splend.NewOutgoingFromDB(*outgoing.ID, dbh)

	assert.True(t, errors.Is(err, splend.ErrOutgoingUnknown))
}

func TestOutgoing_ToggleSettled(t *testing.T) {
	dbh := config.TestDBH()
	randomOutgoing := test.RandomUserAndOutgoing(dbh)
	outgoing, _ := splend.NewOutgoing(randomOutgoing, dbh)

	assert.Nil(t, outgoing.Settled)

	err := outgoing.ToggleSettled(true)

	assert.Nil(t, err)
	assert.NotNil(t, outgoing.Settled)
}

func TestOutgoing_Update(t *testing.T) {
	dbh := config.TestDBH()
	randomOutgoing := test.RandomUserAndOutgoing(dbh)
	outgoing, _ := splend.NewOutgoing(randomOutgoing, dbh)

	outgoing.Description = "Groceries"
	err := outgoing.Update()

	assert.Nil(t, err)

	updatedOutgoing, _ := splend.NewOutgoingFromDB(*outgoing.ID, dbh)

	assert.Equal(t, updatedOutgoing.Description, "Groceries")
}

func TestOutgoing_UpdateTags(t *testing.T) {
	dbh := config.TestDBH()
	randomOutgoing := test.RandomUserAndOutgoing(dbh)
	outgoing, _ := splend.NewOutgoing(randomOutgoing, dbh)

	newTags := []string{"non-discretionary"}
	err := outgoing.UpdateTags(newTags)
	assert.Nil(t, err)

	updatedOutgoing, _ := splend.NewOutgoingFromDB(*outgoing.ID, dbh)

	assert.Equal(t, newTags, updatedOutgoing.Tags)
}
