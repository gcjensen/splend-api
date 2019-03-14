package splend

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type User struct {
	dbh       *sql.DB
	ID        *int    `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Colour    *string `json:"colour"`
	Partner   *User   `json:"partner"`
	CoupleID  *int    `json:"-"`
	IconLink  *string `json:"iconLink"`
}

func NewUser(user *User, dbh *sql.DB) (*User, error) {
	self := user
	self.dbh = dbh

	err := self.getInsertDetails()

	return self, err
}

func NewUserFromDB(id int, dbh *sql.DB) (*User, error) {
	statement := fmt.Sprintf(`
        SELECT email
        FROM users
        WHERE id=%d`, id)

	self := &User{dbh: dbh}
	err := dbh.QueryRow(statement).Scan(&self.Email)

	if err != nil {
		return nil, errors.New("Unknown user")
	}

	err = self.getUser()

	return self, err
}

func (self *User) AddOutgoing(o *Outgoing) error {
	o.Spender = *self.ID
	_, err := NewOutgoing(o, self.dbh)
	return err
}

func (self *User) GetOutgoings() ([]Outgoing, error) {
	ids := []string{strconv.Itoa(*self.ID)}
	if self.Partner.ID != nil {
		ids = append(ids, strconv.Itoa(*self.Partner.ID))
	}
	statement := fmt.Sprintf(`
		SELECT o.id, description, amount, owed, spender_id, c.name, settled,
		timestamp
		FROM outgoings o
		JOIN categories c ON o.category_id=c.id
		WHERE spender_id IN (%s)
		ORDER BY o.timestamp DESC`,
		strings.Join(ids, ","))

	rows, err := self.dbh.Query(statement)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var outgoings []Outgoing

	for rows.Next() {
		var o Outgoing
		if err := rows.Scan(&o.ID, &o.Description, &o.Amount, &o.Owed,
			&o.Spender, &o.Category, &o.Settled, &o.Timestamp); err != nil {
			return nil, err
		}
		outgoings = append(outgoings, o)
	}

	return outgoings, err
}

/************************** Private Implementation ****************************/

func (self *User) addCouple() int {
	statement := fmt.Sprintf(
		`INSERT INTO couples (joining_date) VALUES ("2018-01-01")`,
	)

	self.dbh.Exec(statement)

	var id int
	self.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

func (self *User) getInsertDetails() error {
	err := self.getUser()
	if err != nil {
		if self.CoupleID == nil {
			coupleID := self.addCouple()
			self.CoupleID = &coupleID
		}
		statement := fmt.Sprintf(`
			INSERT INTO users
			(first_name, last_name, email, couple_id, colour)
			VALUES ("%s", "%s", "%s", %d, "%s")`,
			self.FirstName, self.LastName, self.Email, *self.CoupleID,
			*self.Colour)

		_, err = self.dbh.Exec(statement)

		self.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&self.ID)
		self.getPartner()
	}

	return nil
}

func (self *User) getUser() error {
	statement := fmt.Sprintf(`
        SELECT id, first_name, last_name, couple_id, colour, icon_link
        FROM users
        WHERE email="%s"`, self.Email)

	err := self.dbh.QueryRow(statement).Scan(
		&self.ID,
		&self.FirstName,
		&self.LastName,
		&self.CoupleID,
		&self.Colour,
		&self.IconLink,
	)

	if err != nil {
		return errors.New("Unknown user")
	}

	err = self.getPartner()
	if err != nil {
		return err
	}

	return err
}

func (self *User) getPartner() error {
	if self.CoupleID == nil {
		return errors.New("No partner")
	}

	statement := fmt.Sprintf(`
        SELECT id, first_name, last_name, email, colour, couple_id, icon_link
		FROM users
		WHERE couple_id = %d AND id != %d`,
		*self.CoupleID, *self.ID)

	partner := &User{}
	err := self.dbh.QueryRow(statement).Scan(
		&partner.ID,
		&partner.FirstName,
		&partner.LastName,
		&partner.Email,
		&partner.Colour,
		&partner.CoupleID,
		&partner.IconLink,
	)
	self.Partner = partner

	if err != sql.ErrNoRows {
		return err
	}
	return nil
}
