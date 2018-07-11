package user

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gcjensen/settle-api/outgoing"
)

type User struct {
	dbh       *sql.DB
	ID        *int    `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Colour    *string `json:"colour"`
	Partner   struct {
		ID     int     `json:"id"`
		Name   string  `json:"name"`
		Colour *string `json:"colour"`
	} `json:"partner"`
}

func New(email string, dbh *sql.DB) (*User, error) {

	// TODO: Parse email to check validity

	self := &User{Email: email}
	self.dbh = dbh

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

func (self *User) AddOutgoing(o outgoing.Outgoing) error {

	statement := fmt.Sprintf(
		`SELECT id FROM categories WHERE name = "%s"`,
		o.Category,
	)

	var categoryID int
	err := self.dbh.QueryRow(statement).Scan(&categoryID)

	if err != nil {
		return errors.New("Invalid category")
	}

	statement = fmt.Sprintf(`
		INSERT INTO outgoings
		(description, amount, owed, spender_id, category_id, settled, timestamp)
		VALUES ("%s", %f, %f, %d, %d, NULL, NOW())`,
		o.Description, o.Amount, o.Owed, o.Spender, categoryID,
	)

	_, err = self.dbh.Exec(statement)

	return err
}

func (self *User) GetOutgoings() ([]outgoing.Outgoing, error) {
	statement := fmt.Sprintf(`
		SELECT o.id, description, amount, owed, spender_id, c.name, settled, timestamp
		FROM outgoings o
		JOIN categories c ON o.category_id=c.id
		WHERE spender_id IN (%d, %d)
		ORDER BY o.timestamp DESC`,
		*self.ID, self.Partner.ID)

	rows, err := self.dbh.Query(statement)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var outgoings []outgoing.Outgoing

	for rows.Next() {
		var o outgoing.Outgoing
		if err := rows.Scan(&o.ID, &o.Description, &o.Amount, &o.Owed,
			&o.Spender, &o.Category, &o.Settled, &o.Timestamp); err != nil {
			return nil, err
		}
		outgoings = append(outgoings, o)
	}

	return outgoings, err
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
        SELECT id, first_name, last_name, couple_id, colour
        FROM users
        WHERE email="%s"`, self.Email)

	var coupleID int
	err := dbh.QueryRow(statement).Scan(
		&self.ID,
		&self.FirstName,
		&self.LastName,
		&coupleID,
		&self.Colour,
	)

	if err != nil {
		return errors.New("Unknown user")
	}

	err = self.getPartner(coupleID, dbh)
	if err != nil {
		return err
	}

	return err
}

func (self *User) getPartner(coupleID int, dbh *sql.DB) error {
	statement := fmt.Sprintf(`
		SELECT id, first_name, colour
		FROM users
		WHERE couple_id = %d AND id != %d`,
		coupleID, *self.ID)

	err := dbh.QueryRow(statement).Scan(
		&self.Partner.ID,
		&self.Partner.Name,
		&self.Partner.Colour,
	)

	if err != sql.ErrNoRows {
		return err
	}
	return nil
}
