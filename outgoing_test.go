package splend

import (
	"database/sql"
	"errors"
	"math/rand"
	"testing"

	"github.com/gcjensen/splend-api/config"
	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	dbh := config.TestDBH()

	randomOutgoing := randomUserAndOutgoing(dbh)
	outgoing, _ := NewOutgoing(randomOutgoing, dbh)
	outgoing.dbh = dbh

	outgoingFromDB, err := NewOutgoingFromDB(*outgoing.ID, dbh)
	outgoing.Timestamp = outgoingFromDB.Timestamp

	assert.Nil(t, err)
	assert.Equal(t, outgoing, outgoingFromDB)

	_, err = NewOutgoingFromDB(10000, dbh)
	assert.NotNil(t, err)
}

func TestDelete(t *testing.T) {
	dbh := config.TestDBH()
	randomOutgoing := randomUserAndOutgoing(dbh)
	outgoing, _ := NewOutgoing(randomOutgoing, dbh)
	outgoing.dbh = dbh

	err := outgoing.Delete()
	assert.Nil(t, err)

	_, err = NewOutgoingFromDB(*outgoing.ID, dbh)

	assert.True(t, errors.Is(err, ErrOutgoingUnknown))

	// Make sure we can also delete outgoings that exist in amex_transactions
	outgoing, _ = NewOutgoing(randomUserAndOutgoing(dbh), dbh)
	outgoing.dbh = dbh

	statement, _ := dbh.Prepare(`
		INSERT INTO amex_transactions
		(amex_id, outgoing_id)
		VALUES ("aaaaaaaaaaaaaaaaaaaaaaa", ?)
	`)
	_, _ = statement.Exec(outgoing.ID)

	err = outgoing.Delete()
	assert.Nil(t, err)

	_, err = NewOutgoingFromDB(*outgoing.ID, dbh)

	assert.True(t, errors.Is(err, ErrOutgoingUnknown))
}

func TestToggleSettled(t *testing.T) {
	dbh := config.TestDBH()
	randomOutgoing := randomUserAndOutgoing(dbh)
	outgoing, _ := NewOutgoing(randomOutgoing, dbh)
	outgoing.dbh = dbh

	assert.Nil(t, outgoing.Settled)

	err := outgoing.ToggleSettled(true)

	assert.Nil(t, err)
	assert.NotNil(t, outgoing.Settled)
}

func TestUpdated(t *testing.T) {
	dbh := config.TestDBH()
	randomOutgoing := randomUserAndOutgoing(dbh)
	outgoing, _ := NewOutgoing(randomOutgoing, dbh)
	outgoing.dbh = dbh

	outgoing.Description = "Groceries"
	err := outgoing.Update()

	assert.Nil(t, err)

	updatedOutgoing, _ := NewOutgoingFromDB(*outgoing.ID, dbh)

	assert.Equal(t, updatedOutgoing.Description, "Groceries")
}

func randomUserAndOutgoing(dbh *sql.DB) *Outgoing {
	statement, _ := dbh.Prepare(`
		INSERT INTO users
		(first_name, last_name, email, sha256)
		VALUES (?, ?, ?, "")
	`)

	_, _ = statement.Exec(fake.FirstName(), fake.LastName(), fake.EmailAddress())

	var spenderID int
	_ = dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&spenderID)

	outgoing := randomOutgoing()
	outgoing.Spender = spenderID

	return outgoing
}

func randomOutgoing() *Outgoing {
	amount := rand.Intn(100)

	return &Outgoing{
		Description: fake.ProductName(),
		Amount:      amount,
		Owed:        amount / 2,
		Category:    fake.Product(),
	}
}
