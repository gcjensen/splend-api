package http

import (
	"github.com/gcjensen/splend"
	"github.com/icrowley/fake"
)

func randomUser() *splend.User {
	colour := "FFFFFF"
	return &splend.User{
		FirstName: fake.FirstName(),
		LastName:  fake.LastName(),
		Email:     fake.EmailAddress(),
		Colour:    &colour,
	}
}

func randomOutgoing() *splend.Outgoing {
	return &splend.Outgoing{
		Description: fake.ProductName(),
		Amount:      10.22,
		Owed:        5.11,
		Category:    fake.Product(),
	}
}
