package test

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/gcjensen/splend-api/splend"
	"github.com/icrowley/fake"
)

func RandomSha256() string {
	return fmt.Sprintf(
		"%x",
		sha256.Sum256([]byte(fake.Digits())),
	)
}

func RandomUser() *splend.User {
	colour := "FFFFFF"

	return &splend.User{
		FirstName: fake.FirstName(),
		LastName:  fake.LastName(),
		Email:     fake.EmailAddress(),
		Colour:    &colour,
	}
}

func RandomUserAndOutgoing(dbh *sql.DB) *splend.Outgoing {
	statement, _ := dbh.Prepare(`
		INSERT INTO users
		(first_name, last_name, email, sha256)
		VALUES (?, ?, ?, "")
	`)
	defer statement.Close()

	_, _ = statement.Exec(fake.FirstName(), fake.LastName(), fake.EmailAddress())

	var spenderID int
	_ = dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&spenderID)

	outgoing := RandomOutgoing()
	outgoing.Spender = spenderID

	return outgoing
}

func RandomOutgoing() *splend.Outgoing {
	amount := rand.Intn(100)

	return &splend.Outgoing{
		Description: fake.ProductName(),
		Amount:      amount,
		Owed:        amount / 2,
		Category:    fake.Product(),
	}
}
