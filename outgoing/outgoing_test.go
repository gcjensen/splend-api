package outgoing

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
)

func TestNew(t *testing.T) {
	dbh := config.TestDBH()

	randomOutgoing := randomOutgoing(dbh)
	outgoing, err := New(randomOutgoing, dbh)
	outgoing.dbh = dbh

	outgoingFromDB, err := NewFromDB(*outgoing.ID, dbh)
	outgoing.Timestamp = outgoingFromDB.Timestamp

	assert.Nil(t, err)
	assert.Equal(t, outgoing, outgoingFromDB)

	outgoing, err = NewFromDB(10000, dbh)
	assert.NotNil(t, err)
}

func TestDelete(t *testing.T) {
	dbh := config.TestDBH()
	randomOutgoing := randomOutgoing(dbh)
	outgoing, err := New(randomOutgoing, dbh)
	outgoing.dbh = dbh

	err = outgoing.Delete()
	assert.Nil(t, err)

	_, err = NewFromDB(*outgoing.ID, dbh)

	assert.Equal(t, err, errors.New("Unknown outgoing"))
}

func TestToggleSettled(t *testing.T) {
	dbh := config.TestDBH()
	randomOutgoing := randomOutgoing(dbh)
	outgoing, err := New(randomOutgoing, dbh)
	outgoing.dbh = dbh

	assert.Nil(t, outgoing.Settled)

	err = outgoing.ToggleSettled(true)

	assert.Nil(t, err)
	assert.NotNil(t, outgoing.Settled)
}

/************************** Private Implementation ****************************/

func randomOutgoing(dbh *sql.DB) *Outgoing {
	statement := fmt.Sprintf(`
		INSERT INTO users
		(first_name, last_name, email)
		VALUES ("%s", "%s", "%s")`,
		fake.FirstName(), fake.LastName(), fake.EmailAddress())

	dbh.Exec(statement)

	var spenderID int
	dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&spenderID)

	amount := math.Ceil(rand.Float64()*100) / 100
	return &Outgoing{
		Description: fake.ProductName(),
		Amount:      amount,
		Owed:        amount / 2,
		Category:    fake.Product(),
		Spender:     spenderID,
	}
}
