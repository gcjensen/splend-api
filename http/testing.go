package http

import (
	"crypto/sha256"
	"fmt"
	"math/rand"

	"github.com/gcjensen/splend-api"
	"github.com/icrowley/fake"
)

func randomOutgoing() *splend.Outgoing {
	amount := rand.Intn(100)

	return &splend.Outgoing{
		Description: fake.ProductName(),
		Amount:      amount,
		Owed:        amount / 2,
		Category:    fake.Product(),
	}
}

func randomSha256() string {
	return fmt.Sprintf(
		"%x",
		sha256.Sum256([]byte(fake.Digits())),
	)
}

func randomUser() *splend.User {
	colour := "FFFFFF"

	return &splend.User{
		FirstName: fake.FirstName(),
		LastName:  fake.LastName(),
		Email:     fake.EmailAddress(),
		Colour:    &colour,
	}
}
