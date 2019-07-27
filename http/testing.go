package http

import (
	"crypto/sha256"
	"fmt"
	"github.com/gcjensen/splend"
	"github.com/icrowley/fake"
)

func randomOutgoing() *splend.Outgoing {
	return &splend.Outgoing{
		Description: fake.ProductName(),
		Amount:      10.22,
		Owed:        5.11,
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
