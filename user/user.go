package user

import (
	"database/sql"
	"errors"
	"fmt"
)

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Partner   string `json:"partner"`
}

func New(email string, dbh *sql.DB) (*User, error) {

	// TODO: Parse email to check validity

	self := &User{Email: email}

	err := self.getInsertDetails(dbh)

	return self, err
}

func NewFromDB(id int, dbh *sql.DB) (*User, error) {
	statement := fmt.Sprintf(`
        SELECT email
        FROM users
        WHERE id=%d`, id)

	var email string
	err := dbh.QueryRow(statement).Scan(&email)

	if err != nil {
		return nil, errors.New("Unknown user")
	}

	return New(email, dbh)
}

/************************** Private Implementation ****************************/

func (self *User) getInsertDetails(dbh *sql.DB) error {
	err := self.getUser(dbh)
	if err != nil {
		// TODO implement user creation
		return errors.New("User creation not yet implemented")
	}

	return nil
}

func (self *User) getUser(dbh *sql.DB) error {
	statement := fmt.Sprintf(`
        SELECT id, first_name, last_name, couple_id
        FROM users
        WHERE email="%s"`, self.Email)

	var coupleID int
	err := dbh.QueryRow(statement).Scan(
		&self.ID,
		&self.FirstName,
		&self.LastName,
		&coupleID,
	)

	if err != nil {
		return errors.New("Unknown user")
	}

	err = self.getPartnerName(coupleID, dbh)
	if err != nil {
		return err
	}

	return err
}

func (self *User) getPartnerName(coupleID int, dbh *sql.DB) error {
	statement := fmt.Sprintf(`
		SELECT first_name
		FROM users
		WHERE couple_id = %d AND id != %d`,
		coupleID, self.ID)

	err := dbh.QueryRow(statement).Scan(&self.Partner)
	if err != sql.ErrNoRows {
		return err
	}
	return nil
}
