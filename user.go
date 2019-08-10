package splend

import (
	"database/sql"
	"errors"
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

func NewUser(user *User, sha256 string, dbh *sql.DB) (*User, error) {
	self := user
	self.dbh = dbh

	err := self.getInsertDetails(sha256)

	return self, err
}

func NewUserFromDB(id int, dbh *sql.DB) (*User, error) {
	self := &User{dbh: dbh}
	err := dbh.QueryRow(
		`SELECT email FROM users WHERE id = ?`, id,
	).Scan(&self.Email)

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
	var partnerID int
	if self.Partner.ID != nil {
		partnerID = *self.Partner.ID
	}

	query := `
		SELECT o.id, description, amount, owed, spender_id, c.name, settled,
		timestamp
		FROM outgoings o
		JOIN categories c ON o.category_id=c.id
		WHERE spender_id = ? OR (spender_id = ? AND owed > 0)
		ORDER BY o.timestamp DESC
	`

	rows, err := self.dbh.Query(query, self.ID, partnerID)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

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
	self.dbh.Exec(`
		INSERT INTO couples (joining_date) VALUES (NOW())
	`)

	var id int
	self.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

func (self *User) getInsertDetails(sha256 string) error {
	err := self.getUser()
	if err != nil {
		if self.CoupleID == nil {
			coupleID := self.addCouple()
			self.CoupleID = &coupleID
		}
		statement, _ := self.dbh.Prepare(`
			INSERT INTO users
			(first_name, last_name, email, couple_id, colour, sha256)
			VALUES (?, ?, ?, ?, ?, ?)
		`)

		_, err = statement.Exec(
			self.FirstName, self.LastName, self.Email, *self.CoupleID,
			*self.Colour, sha256,
		)

		self.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&self.ID)
		self.getPartner()
	}

	return nil
}

func (self *User) getUser() error {
	query := `
        SELECT id, first_name, last_name, couple_id, colour, icon_link
        FROM users
        WHERE email= ?
	`

	err := self.dbh.QueryRow(query, self.Email).Scan(
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

	query := `
        SELECT id, first_name, last_name, email, colour, couple_id, icon_link
		FROM users
		WHERE couple_id = ? AND id != ?
	`

	partner := &User{}
	err := self.dbh.QueryRow(query, self.CoupleID, self.ID).Scan(
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
