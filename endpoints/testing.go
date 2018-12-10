package endpoints

import (
	"github.com/gcjensen/settle-api/outgoing"
	"github.com/gcjensen/settle-api/user"
	"github.com/icrowley/fake"
)

func randomUser() *user.User {
	colour := "FFFFFF"
	return &user.User{
		FirstName: fake.FirstName(),
		LastName:  fake.LastName(),
		Email:     fake.EmailAddress(),
		Colour:    &colour,
	}
}

func randomOutgoing() *outgoing.Outgoing {
	return &outgoing.Outgoing{
		Description: fake.ProductName(),
		Amount:      10.22,
		Owed:        5.11,
		Category:    fake.Product(),
	}
}
