package test

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

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

func RandomCouple(dbh *sql.DB) (*splend.User, *splend.User) {
	user, _ := splend.NewUser(RandomUser(), RandomSha256(), dbh)
	randomPartner := RandomUser()
	randomPartner.CoupleID = user.CoupleID
	partner, _ := splend.NewUser(randomPartner, RandomSha256(), dbh)
	user.Partner = partner

	return user, partner
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
	now := time.Date(2024, 05, 07, 21, 17, 14, 0, time.UTC)

	return &splend.Outgoing{
		Description: fake.ProductName(),
		Amount:      amount,
		Owed:        amount / 2,
		Category:    fake.Product(),
		Tags:        []string{fake.Product(), fake.Product()},
		Timestamp:   &now,
	}
}
